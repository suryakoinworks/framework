package dics

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/KejawenLab/bima/v3"
	"github.com/KejawenLab/bima/v3/configs"
	"github.com/KejawenLab/bima/v3/drivers"
	"github.com/KejawenLab/bima/v3/events"
	"github.com/KejawenLab/bima/v3/generators"
	"github.com/KejawenLab/bima/v3/handlers"
	"github.com/KejawenLab/bima/v3/interfaces"
	"github.com/KejawenLab/bima/v3/listeners/creates"
	"github.com/KejawenLab/bima/v3/listeners/deletes"
	filters "github.com/KejawenLab/bima/v3/listeners/paginations"
	"github.com/KejawenLab/bima/v3/listeners/updates"
	"github.com/KejawenLab/bima/v3/loggers"
	"github.com/KejawenLab/bima/v3/middlewares"
	"github.com/KejawenLab/bima/v3/models"
	paginations "github.com/KejawenLab/bima/v3/paginations"
	"github.com/KejawenLab/bima/v3/paginations/adapter"
	"github.com/KejawenLab/bima/v3/repositories"
	"github.com/KejawenLab/bima/v3/routers"
	"github.com/KejawenLab/bima/v3/routes"
	"github.com/KejawenLab/bima/v3/utils"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/fatih/color"
	"github.com/gertd/go-pluralize"
	"github.com/kamva/mgm/v3"
	"github.com/olivere/elastic/v7"
	"github.com/sarulabs/dingo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

var Container = []dingo.Def{
	{
		Name:  "bima:config:template",
		Build: (*generators.Template)(nil),
	},
	{
		Name:  "bima:template:module",
		Build: (*generators.ModuleTemplate)(nil),
	},
	{
		Name:  "bima:template:field",
		Build: (*generators.FieldTemplate)(nil),
	},
	{
		Name:  "bima:config",
		Build: (*configs.Env)(nil),
	},
	{
		Name: "bima:module:generator",
		Build: func(
			dic generators.Generator,
			model generators.Generator,
			module generators.Generator,
			proto generators.Generator,
			provider generators.Generator,
			server generators.Generator,
			swagger generators.Generator,
			env *configs.Env,
			template *generators.Template,
		) (*generators.Factory, error) {
			return &generators.Factory{
				ApiVersion: env.ApiVersion,
				Driver:     env.Db.Driver,
				Pluralizer: pluralize.NewClient(),
				Template:   template,
				Generators: []generators.Generator{dic, model, module, proto, provider, server, swagger},
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:generator:dic"),
			"1": dingo.Service("bima:generator:model"),
			"2": dingo.Service("bima:generator:module"),
			"3": dingo.Service("bima:generator:proto"),
			"4": dingo.Service("bima:generator:provider"),
			"5": dingo.Service("bima:generator:server"),
			"6": dingo.Service("bima:generator:swagger"),
			"7": dingo.Service("bima:config"),
			"8": dingo.Service("bima:config:template"),
		},
	},
	{
		Name:  "bima:generator:dic",
		Build: (*generators.Dic)(nil),
	},
	{
		Name:  "bima:generator:model",
		Build: (*generators.Model)(nil),
	},
	{
		Name:  "bima:generator:module",
		Build: (*generators.Module)(nil),
	},
	{
		Name:  "bima:generator:proto",
		Build: (*generators.Proto)(nil),
	},
	{
		Name:  "bima:generator:provider",
		Build: (*generators.Provider)(nil),
	},
	{
		Name:  "bima:generator:server",
		Build: (*generators.Server)(nil),
	},
	{
		Name:  "bima:generator:swagger",
		Build: (*generators.Swagger)(nil),
	},
	{
		Name:  "bima:database:driver:mysql",
		Build: (*drivers.Mysql)(nil),
	},
	{
		Name:  "bima:database:driver:postgresql",
		Build: (*drivers.PostgreSql)(nil),
	},
	{
		Name: "bima:application",
		Build: func(
			database interfaces.Application,
			elasticsearch interfaces.Application,
			grpc interfaces.Application,
			queue interfaces.Application,
			rest interfaces.Application,
		) (*interfaces.Factory, error) {
			return &interfaces.Factory{
				Applications: []interfaces.Application{database, elasticsearch, grpc, queue, rest},
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:interface:database"),
			"1": dingo.Service("bima:interface:elasticsearch"),
			"2": dingo.Service("bima:interface:grpc"),
			"3": dingo.Service("bima:interface:queue"),
			"4": dingo.Service("bima:interface:rest"),
		},
	},
	{
		Name:  "bima:event:dispatcher",
		Build: (*events.Dispatcher)(nil),
	},
	{
		Name: "bima:middleware:factory",
		Build: func(env *configs.Env, logger *loggers.Logger) (*middlewares.Factory, error) {
			middleware := middlewares.Factory{
				Debug:  env.Debug,
				Logger: logger,
			}
			middleware.Add(&middlewares.Header{})

			return &middleware, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:handler:logger"),
		},
	},
	{
		Name:  "bima:logger:extension",
		Build: (*loggers.LoggerExtension)(nil),
	},
	{
		Name: "bima:connection:database",
		Build: func(env *configs.Env, mysql drivers.Driver, postgresql drivers.Driver) (*gorm.DB, error) {
			util := color.New(color.FgCyan, color.Bold)
			var db drivers.Driver

			util.Print("✓ ")
			fmt.Print("Database configured using ")
			util.Print(env.Db.Driver)
			fmt.Println(" driver")

			switch env.Db.Driver {
			case "mysql":
				db = mysql
			case "postgresql":
				db = postgresql
			case "mongo":
				var dsn bytes.Buffer

				dsn.WriteString("mongodb://")
				dsn.WriteString(env.Db.User)
				dsn.WriteString(":")
				dsn.WriteString(env.Db.Password)
				dsn.WriteString("@")
				dsn.WriteString(env.Db.Host)
				dsn.WriteString(":")
				dsn.WriteString(strconv.Itoa(env.Db.Port))

				err := mgm.SetDefaultConfig(nil, env.Db.Name, options.Client().ApplyURI(dsn.String()).SetMonitor(&event.CommandMonitor{
					Started: func(_ context.Context, evt *event.CommandStartedEvent) {
						log.Print(evt.Command)
					},
				}))
				if err != nil {
					dsn.Reset()
					dsn.WriteString("mongodb://")
					dsn.WriteString(env.Db.Host)

					err = mgm.SetDefaultConfig(nil, env.Db.Name, options.Client().ApplyURI(dsn.String()).SetMonitor(&event.CommandMonitor{
						Started: func(_ context.Context, evt *event.CommandStartedEvent) {
							log.Print(evt.Command)
						},
					}))
				}

				return nil, err
			default:
				return nil, errors.New("Unknown database driver")
			}

			return db.Connect(
				env.Db.Host,
				env.Db.Port,
				env.Db.User,
				env.Db.Password,
				env.Db.Name,
				env.Debug,
			), nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:database:driver:mysql"),
			"2": dingo.Service("bima:database:driver:postgresql"),
		},
	},
	{
		Name: "bima:connection:elasticsearch",
		Build: func(env *configs.Env) (*elastic.Client, error) {
			var dsn bytes.Buffer

			dsn.WriteString(env.Elasticsearch.Host)
			dsn.WriteString(":")
			dsn.WriteString(strconv.Itoa(env.Elasticsearch.Port))

			client, err := elastic.NewClient(
				elastic.SetURL(dsn.String()),
				elastic.SetSniff(false),
				elastic.SetHealthcheck(false),
				elastic.SetGzip(true),
			)

			if err != nil {
				return nil, nil
			}

			color.New(color.FgCyan, color.Bold).Print("✓ ")
			fmt.Println("Elasticsearch configured")

			return client, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
		},
	},
	{
		Name: "bima:listener:create:elasticsearch",
		Build: func(env *configs.Env, client *elastic.Client) (*creates.Elasticsearch, error) {
			return &creates.Elasticsearch{
				Service:       env.Service.ConnonicalName,
				Elasticsearch: client,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:connection:elasticsearch"),
		},
	},
	{
		Name: "bima:listener:update:elasticsearch",
		Build: func(env *configs.Env, client *elastic.Client) (*updates.Elasticsearch, error) {
			return &updates.Elasticsearch{
				Service:       env.Service.ConnonicalName,
				Elasticsearch: client,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:connection:elasticsearch"),
		},
	},
	{
		Name: "bima:listener:delete:elasticsearch",
		Build: func(env *configs.Env, client *elastic.Client) (*deletes.Elasticsearch, error) {
			return &deletes.Elasticsearch{
				Service:       env.Service.ConnonicalName,
				Elasticsearch: client,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:connection:elasticsearch"),
		},
	},
	{
		Name:  "bima:listener:filter:elasticsearch",
		Build: (*filters.ElasticsearchFilter)(nil),
	},
	{
		Name:  "bima:listener:filter:gorm",
		Build: (*filters.GormFilter)(nil),
	},
	{
		Name:  "bima:listener:filter:mongo",
		Build: (*filters.MongoDbFilter)(nil),
	},
	{
		Name:  "bima:interface:database",
		Build: (*interfaces.Database)(nil),
	},
	{
		Name:  "bima:interface:elasticsearch",
		Build: (*interfaces.Elasticsearch)(nil),
	},
	{
		Name: "bima:interface:grpc",
		Build: func(env *configs.Env) (*interfaces.GRpc, error) {
			return &interfaces.GRpc{
				GRpcPort: env.RpcPort,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
		},
	},
	{
		Name:  "bima:interface:queue",
		Build: (*interfaces.Queue)(nil),
	},
	{
		Name: "bima:interface:rest",
		Build: func(
			env *configs.Env,
			middleware *middlewares.Factory,
			router *routers.Factory,
		) (*interfaces.Rest, error) {
			return &interfaces.Rest{
				GRpcPort:   env.RpcPort,
				HttpPort:   env.HttpPort,
				Middleware: middleware,
				Router:     router,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:middleware:factory"),
			"2": dingo.Service("bima:handler:router"),
		},
	},
	{
		Name: "bima:handler:logger",
		Build: func(
			env *configs.Env,
			extension *loggers.LoggerExtension,
		) (*loggers.Logger, error) {
			logger := logrus.New()
			if env.Debug {
				logger.SetLevel(logrus.DebugLevel)
			}

			logger.SetFormatter(&logrus.TextFormatter{
				FullTimestamp: true,
			})
			for _, e := range extension.Extensions {
				logger.AddHook(e)
			}

			return &loggers.Logger{
				Verbose: env.Debug,
				Service: env.Service,
				Logger:  logger,
				Data:    logrus.Fields{},
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:logger:extension"),
		},
	},
	{
		Name:  "bima:handler:messager",
		Build: (*handlers.Messenger)(nil),
		Params: dingo.Params{
			"Logger":    dingo.Service("bima:handler:logger"),
			"Publisher": dingo.Service("bima:message:publisher"),
			"Consumer":  dingo.Service("bima:message:consumer"),
		},
	},
	{
		Name: "bima:handler:router",
		Build: func(gateway routers.Router, mux routers.Router) (*routers.Factory, error) {
			return &routers.Factory{
				Routers: []routers.Router{
					gateway,
					mux,
				},
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:router:gateway"),
			"1": dingo.Service("bima:router:mux"),
		},
	},
	{
		Name:  "bima:middleware:auth",
		Build: (*middlewares.Auth)(nil),
		Params: dingo.Params{
			"Env":    dingo.Service("bima:config"),
			"Logger": dingo.Service("bima:handler:logger"),
		},
	},
	{
		Name: "bima:middleware:requestid",
		Build: func(env *configs.Env, logger *loggers.Logger) (*middlewares.RequestID, error) {
			return &middlewares.RequestID{
				Logger:          logger,
				RequestIDHeader: env.RequestIDHeader,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:handler:logger"),
		},
	},
	{
		Name: "bima:router:mux",
		Build: func(
			apiDoc routes.Route,
			apiDocRedirection routes.Route,
			health routes.Route,
		) (*routers.MuxRouter, error) {
			routers := routers.MuxRouter{}
			routers.Register([]routes.Route{apiDoc, apiDocRedirection, health})

			return &routers, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:route:api-doc"),
			"1": dingo.Service("bima:route:api-doc-redirect"),
			"2": dingo.Service("bima:route:health"),
		},
	},
	{
		Name:  "bima:router:gateway",
		Build: (*routers.GRpcGateway)(nil),
	},
	{
		Name: "bima:route:api-doc",
		Build: func(env *configs.Env) (*routes.ApiDoc, error) {
			return &routes.ApiDoc{
				Debug: env.Debug,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
		},
	},
	{
		Name:  "bima:route:api-doc-redirect",
		Build: (*routes.ApiDocRedirect)(nil),
	},
	{
		Name:  "bima:route:health",
		Build: (*routes.Health)(nil),
		Params: dingo.Params{
			"Logger": dingo.Service("bima:handler:logger"),
		},
	},
	{
		Name: "bima:message:config",
		Build: func(env *configs.Env) (amqp.Config, error) {
			color.New(color.FgCyan, color.Bold).Print("✓ ")
			fmt.Println("Pub/Sub configured")

			var dsn bytes.Buffer

			dsn.WriteString("amqp://")
			dsn.WriteString(env.Amqp.User)
			dsn.WriteString(":")
			dsn.WriteString(env.Amqp.Password)
			dsn.WriteString("@")
			dsn.WriteString(env.Amqp.Host)
			dsn.WriteString(":")
			dsn.WriteString(strconv.Itoa(env.Amqp.Port))

			return amqp.NewDurableQueueConfig(dsn.String()), nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
		},
	},
	{
		Name: "bima:message:publisher",
		Build: func(env *configs.Env, config amqp.Config) (*amqp.Publisher, error) {
			publisher, err := amqp.NewPublisher(config, watermill.NewStdLogger(env.Debug, env.Debug))
			if err != nil {
				return nil, err
			}

			return publisher, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:message:config"),
		},
	},
	{
		Name: "bima:message:consumer",
		Build: func(env *configs.Env, config amqp.Config) (*amqp.Subscriber, error) {
			consumer, err := amqp.NewSubscriber(config, watermill.NewStdLogger(env.Debug, env.Debug))
			if err != nil {
				return nil, err
			}

			return consumer, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:message:config"),
		},
	},
	{
		Name:  "bima:pagination:paginator",
		Build: (*paginations.Pagination)(nil),
	},
	{
		Name:  "bima:pagination:adapter:gorm",
		Build: (*adapter.GormAdapter)(nil),
		Params: dingo.Params{
			"Logger":     dingo.Service("bima:handler:logger"),
			"Database":   dingo.Service("bima:connection:database"),
			"Dispatcher": dingo.Service("bima:event:dispatcher"),
		},
	},
	{
		Name: "bima:pagination:adapter:elasticsearch",
		Build: func(env *configs.Env, logger *loggers.Logger, client *elastic.Client, dispatcher *events.Dispatcher) (*adapter.ElasticsearchAdapter, error) {
			return &adapter.ElasticsearchAdapter{
				Service:    env.Service.ConnonicalName,
				Logger:     logger,
				Client:     client,
				Dispatcher: dispatcher,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:handler:logger"),
			"2": dingo.Service("bima:connection:elasticsearch"),
			"3": dingo.Service("bima:event:dispatcher"),
		},
	},
	{
		Name:  "bima:pagination:adapter:mongo",
		Build: (*adapter.MongodbAdapter)(nil),
		Params: dingo.Params{
			"Logger":     dingo.Service("bima:handler:logger"),
			"Dispatcher": dingo.Service("bima:event:dispatcher"),
		},
	},
	{
		Name:  "bima:pagination:request",
		Build: (*paginations.Request)(nil),
	},
	{
		Name:  "bima:service:repository:gorm",
		Build: (*repositories.GormRepository)(nil),
		Params: dingo.Params{
			"Database": dingo.Service("bima:connection:database"),
		},
	},
	{
		Name:  "bima:service:repository:mongo",
		Build: (*repositories.MongoRepository)(nil),
	},
	{
		Name: "bima:cache:memory",
		Build: func(env *configs.Env) (*utils.Cache, error) {
			return utils.NewCache(time.Duration(env.CacheLifetime) * time.Second), nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
		},
	},
	{
		Name: "bima:module",
		Build: func(
			logger *loggers.Logger,
			handler *handlers.Handler,
			cache *utils.Cache,
			paginator *paginations.Pagination,
		) (*bima.Module, error) {
			return &bima.Module{
				Logger:    logger,
				Handler:   handler,
				Cache:     cache,
				Paginator: paginator,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:handler:logger"),
			"1": dingo.Service("bima:handler:handler"),
			"2": dingo.Service("bima:cache:memory"),
			"3": dingo.Service("bima:pagination:paginator"),
		},
	},
	{
		Name: "bima:server",
		Build: func(env *configs.Env, db *gorm.DB) (*bima.Server, error) {
			return &bima.Server{
				Debug:    env.Debug,
				Database: db,
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
			"1": dingo.Service("bima:connection:database"),
		},
	},
	{
		Name: "bima:model",
		Build: func(env *configs.Env) (*bima.GormModel, error) {
			return &bima.GormModel{
				GormBase: models.GormBase{
					Env: env,
				},
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:config"),
		},
	},
}
