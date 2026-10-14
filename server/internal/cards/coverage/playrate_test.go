package coverage

import (
	"strings"
	"testing"
)

const playRateDump = `[
 {"oracle_id":"a","name":"Alpha","edhrec_rank":2},
 {"oracle_id":"a","name":"Alpha","edhrec_rank":9},
 {"oracle_id":"b","name":"Beta","edhrec_rank":1},
 {"oracle_id":"c","name":"Gamma","edhrec_rank":4},
 {"oracle_id":"d","name":"Unranked"},
 {"name":"Front // Back","edhrec_rank":3,"card_faces":[{"oracle_id":"e"},{"oracle_id":"e2"}]}
]`

func TestReadRankedCards(t *testing.T) {
	got, err := ReadRankedCards(strings.NewReader(playRateDump))
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, c := range got {
		ids = append(ids, c.OracleID)
	}
	if want := "b a e c"; strings.Join(ids, " ") != want {
		t.Fatalf("order = %v, want %s", ids, want)
	}
	if got[1].Rank != 2 {
		t.Errorf("a keeps its lowest rank, got %d", got[1].Rank)
	}
	if _, err := ReadRankedCards(strings.NewReader(`{}`)); err == nil {
		t.Error("a non-array dump must be an error")
	}
}

func TestBuildPlayRateReport(t *testing.T) {
	ranked, _ := ReadRankedCards(strings.NewReader(playRateDump))
	rep := BuildPlayRateReport(ranked, map[string]bool{"b": true, "c": true}, []int{4, 2}, 4)
	if len(rep.Rows) != 2 || rep.Rows[0].Top != 2 {
		t.Fatalf("rows not sorted by top: %+v", rep.Rows)
	}
	if r := rep.Rows[0]; r.Ranked != 2 || r.Covered != 1 {
		t.Errorf("top 2 = %+v", r)
	}
	if r := rep.Rows[1]; r.Ranked != 4 || r.Covered != 2 || r.Percent() != 50 {
		t.Errorf("top 4 = %+v", r)
	}
	if len(rep.Missing) != 2 || rep.Missing[0].OracleID != "a" || rep.Missing[1].OracleID != "e" {
		t.Errorf("missing = %+v", rep.Missing)
	}
	if !strings.Contains(rep.String(), "50.0%") {
		t.Errorf("report text lacks the percentage:\n%s", rep)
	}
	if (PlayRateRow{}).Percent() != 0 {
		t.Error("empty row must be 0%")
	}
}

func TestCoveredOracleIDsSkipsBackFaces(t *testing.T) {
	for id := range CoveredOracleIDs() {
		if faceSuffix.MatchString(id) {
			t.Fatalf("%s is a back-face key", id)
		}
	}
}
