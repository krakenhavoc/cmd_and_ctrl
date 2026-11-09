package effects

// Prideful Feastling — Creature — Shapeshifter {2}{W/B} (Lorwyn Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 Lifelink"
func init() {
	Register(Spec{
		OracleID:        "faf2f368-a7e7-4427-bc03-929f294b1e03",
		Name:            "Prideful Feastling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"changeling", "lifelink"},
	})
}
