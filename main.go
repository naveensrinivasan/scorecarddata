package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type scorecard struct {
	Name    string `bigquery:"name"`
	Check   string `bigquery:"check"`
	Score   int    `bigquery:"score"`
	Details string `bigquery:"details"`
	Reason  string `bigquery:"reason"`
}

func main() {
	allchecks := []string{"Code-Review"}
	// sosChecks := []string{"Code-Review","Branch-Protection","Pinned-Dependencies","Dependency-Update-Tool","Fuzzing"}
	if len(os.Args) < 2 {
		log.Fatal("Please provide the path of the go.mod/go.sum location")
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT environment variable must be set.")
	}

	modquery := `
	go list -m -f '{{if not (or  .Main)}}{{.Path}}{{end}}' all \
	| grep "^github" \
	| sort -u \
	| cut -d/ -f1-3 \
	| awk '{print $1}' \
	| tr '\n' ',' | head -c -1
	`
	sql :=
		`
		SELECT
		 distinct(repo.name),
		  c.name as check,
		  c.Score as score,
		  d as details,
		  c.Reason as reason
		FROM` + "`openssf.scorecardcron.scorecard-v2_latest`," +
			`
		  UNNEST(checks) AS c,
		  UNNEST(c.details) d
		WHERE
		 c.name in unnest(@checks) and 
		 c.score < 8 and 
		repo.name IN unnest(@repos)
		group by repo.name,
		  c.name,
		  c.Score,
		  d,
		  reason
		order by score
	`
	// Runs the modquery to generate where clause for the above sql statement.
	c := exec.Command("bash", "-c", fmt.Sprintf("cd %s;", os.Args[1])+modquery)
	data, err := c.Output()
	if err != nil {
		log.Fatal(err)
	}
	parameters := []string{}
	parameters = append(parameters, strings.Split(string(data), ",")...)
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	rows, err := query(ctx, sql, allchecks, parameters, client)
	if err != nil {
		log.Fatal(err)
	}
	result, err := printResults(os.Stdout, rows)
	if err != nil {
		log.Fatal(err)
	}
	j, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(j))
}

func query(ctx context.Context, sql string, allchecks, repos []string, client *bigquery.Client) (*bigquery.RowIterator, error) {
	query := client.Query(sql)
	query.Parameters = []bigquery.QueryParameter{
		{
			Name:  "repos",
			Value: repos,
		},
		{
			Name:  "checks",
			Value: allchecks,
		},
	}
	return query.Read(context.TODO())
}

func printResults(w io.Writer, iter *bigquery.RowIterator) ([]scorecard, error) {
	rows := []scorecard{}
	for {
		var row scorecard
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
