package repo

import (
	"context"
	"github.com/ProjectOort/oort-server/biz/search"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SearchRepo struct {
	_es *elastic.Client
}

const (
	_AsteroidIndex = "oort_server.asteroid"
)

func NewSearchRepo(_es *elastic.Client) *SearchRepo {
	return &SearchRepo{_es: _es}
}

func (x *SearchRepo) SearchAsteroid(ctx context.Context, text string, authorID primitive.ObjectID) ([]*search.Item, error) {
	query := elastic.NewBoolQuery()
	query.Must(
		elastic.NewMatchQuery("author_id", authorID.Hex()),
		elastic.NewQueryStringQuery(text),
	)
	highlight := elastic.NewHighlight()
	highlight.Fields(
		elastic.NewHighlighterField("title").
			PreTags("<em>").
			PostTags("</em>").
			NoMatchSize(50),
		elastic.NewHighlighterField("content").
			PreTags("<em>").
			PostTags("</em>").
			NoMatchSize(50),
	)

	result, err := x._es.Search().
		Index(_AsteroidIndex).
		Query(query).
		Highlight(highlight).
		From(0).
		Size(10).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*search.Item, 0, result.Hits.TotalHits.Value)

	if result.Hits.TotalHits.Value > 0 {
		for _, hit := range result.Hits.Hits {
			var item search.Item
			item.TargetID = hit.Id

			title := hit.Highlight["title"]
			item.Title = title[0]
			content := hit.Highlight["content"]
			item.Content = content
			items = append(items, &item)
		}
	}

	return items, nil
}
