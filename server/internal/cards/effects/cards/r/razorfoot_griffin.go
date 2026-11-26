package cardsr

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Razorfoot Griffin — Creature — Griffin {3}{W}, 2/2 (Magic 2015, #5556):
//
//	"Flying
//	 First strike"
//
// Two printed keywords and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7ecb785-2bfc-4953-b356-db7c27406dbb",
		Name:            "Razorfoot Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
