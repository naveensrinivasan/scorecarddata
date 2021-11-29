Scorecarddata
===
This tool will parse the `go.mod/so.sum` dependencies of a project and fetches the scorecard data for its dependencies. It uses the Google Bigquery data https://github.com/ossf/scorecard#public-data`openssf:scorecardcron.scorecard-v2_latest` to fetch the results


## How to run this?
`go run main.go ~/go/src/github.com/naveensrinivasan/sigstore | jq`

## Prerequisites 

- Google cloud account
- https://cloud.google.com/bigquery/public-data

## Can I get additional checks other than the default?
Yes, the `sql` to fetch the scorecard data can be modified.

## Why should I use this over running the scorecard CLI?
The scorecard CLI would take time to fetch hundreds of repositories, and the GitHub's API will be throttled.

### How are the `go` dependencies parsed?
It is bash goo :face_palm: More on this [explainshell](https://explainshell.com/explain?cmd=go+list+-m+-f+%27%7B%7Bif+not+%28or++.Main%29%7D%7D%7B%7B.Path%7D%7D%7B%7Bend%7D%7D%27+all+++%7C+grep+%22%5Egithub%22+%7C+sort+-u+%7C+cut+-d%2F+-f1-3+%7Cawk+%27%7Bprint+%241%7D%27%7C+sed+%22s%2F%5E%2F%5C%22%2F%3Bs%2F%24%2F%5C%22%2F%22%7C++tr+%27%5Cn%27+%27%2C%27+%7C+head+-c+-1)


