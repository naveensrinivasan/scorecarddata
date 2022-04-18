package bigquery

import (
	"context"
	"fmt"

	bq "cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type Scorecard struct {
	Date string `json:"date" bigquery:"date"`
	Repo struct {
		Name   string `json:"name" bigquery:"name"`
		Commit string `json:"commit" bigquery:"commit"`
	} `json:"repo" bigquery:"repo"`
	Scorecard struct {
		Version string `json:"version" bigquery:"version"`
		Commit  string `json:"commit" bigquery:"commit"`
	} `json:"scorecard" bigquery:"scorecard"`
	Checks []struct {
		Name          string `json:"name" bigquery:"name"`
		Documentation struct {
			Short string `json:"short" bigquery:"short"`
			URL   string `json:"url" bigquery:"url"`
		} `json:"documentation" bigquery:"documentation"`
		Score   int      `json:"score" bigquery:"score"`
		Reason  string   `json:"reason" bigquery:"reason"`
		Details []string `json:"details" bigquery:"details"`
	} `json:"checks" bigquery:"checks"`

	Score float64 `json:"score" bigquery:"score"`
}

type ScorecardOld struct {
	// Name of the GitHub repository
	Name string `bigquery:"name"`
	// Scorecard check
	Check string `bigquery:"check"`
	// Score
	Score int `bigquery:"score"`
	// Details of the scorecard run.
	Details string `bigquery:"details"`
	// The reason for the score.
	Reason string `bigquery:"reason"`
}

// Key is used for ignoring exclude from the results.
// Example Code-Review,github.com/kubernetes/kubernetes
type Key struct {
	Check, Repoistory string
}

type (
	bigquery struct{ project string }
	Bigquery interface {
		// FetchScorecardData fetches scorecard data from Google Bigquery.
		// The checks are the scorecard that are filetred from the bigquery table.
		FetchScorecardData(repos, checks []string, exclusions map[Key]bool) ([]Scorecard, error)
	}
)

// FetchScorecardData fetches scorecard data from Google Bigquery.
// The checks are the scorecard that are filetred from the bigquery table.
func (b bigquery) FetchScorecardData(repos, checks []string, exclusions map[Key]bool) ([]Scorecard, error) {
	fmt.Println(len(repos))
	sql := `SELECT * FROM ` +
		"`openssf.scorecardcron.scorecard-v2_latest `" +
		`WHERE repo.name like '%scorecard%'
limit 1
`
	ctx := context.Background()

	client, err := bq.NewClient(ctx, b.project)
	if err != nil {
		return nil, err
	}

	defer client.Close()

	rows, err := query(ctx, sql, checks, repos, client)
	if err != nil {
		return nil, err
	}

	result, err := fetchResults(rows, exclusions)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// NewBigquery returns the implementaion o
func NewBigquery(projectid string) Bigquery {
	return bigquery{projectid}
}

func query(ctx context.Context, sql string, allchecks, repos []string,
	client *bq.Client,
) (*bq.RowIterator, error) {
	query := client.Query(sql)
	return query.Read(context.TODO())
}

func fetchResults(iter *bq.RowIterator, exclusions map[Key]bool) ([]Scorecard, error) {
	rows := []Scorecard{}
	for {
		var row Scorecard
		err := iter.Next(&row)
		if err == iterator.Done {
			return rows, nil
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
}
