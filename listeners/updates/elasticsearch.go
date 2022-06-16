package updates

import (
	"context"
	"fmt"
	"time"

	"github.com/goccy/go-json"

	"github.com/KejawenLab/bima/v2"
	"github.com/KejawenLab/bima/v2/configs"
	"github.com/KejawenLab/bima/v2/events"
	"github.com/olivere/elastic/v7"
)

type Elasticsearch struct {
	Service       string
	Elasticsearch *elastic.Client
}

func (u *Elasticsearch) Handle(event interface{}) interface{} {
	e := event.(*events.Model)
	m := e.Data.(configs.Model)

	result := make(chan error)
	go func(r chan<- error) {
		query := elastic.NewMatchQuery("Id", e.Id)

		ctx := context.Background()
		result, _ := u.Elasticsearch.Search().Index(fmt.Sprintf("%s_%s", u.Service, m.TableName())).Query(query).Do(ctx)
		if result != nil {
			for _, hit := range result.Hits.Hits {
				u.Elasticsearch.Delete().Index(fmt.Sprintf("%s_%s", u.Service, m.TableName())).Id(hit.Id).Do(ctx)
			}
		}

		data, _ := json.Marshal(e.Data)

		_, err := u.Elasticsearch.Index().Index(fmt.Sprintf("%s_%s", u.Service, m.TableName())).BodyJson(string(data)).Do(ctx)
		r <- err
	}(result)

	go func(r <-chan error) {
		if <-r == nil {
			m.SetSyncedAt(time.Now())
			e.Repository.Update(m)
		}
	}(result)

	return e
}

func (u *Elasticsearch) Listen() string {
	return events.AfterUpdateEvent.String()
}

func (u *Elasticsearch) Priority() int {
	return bima.HighestPriority + 1
}
