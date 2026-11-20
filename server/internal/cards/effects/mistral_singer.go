package effects

// Mistral Singer — Creature — Siren {2}{U}, 2/2 (Core Set 2021, #5511):
//
//	"Flying
//	 Prowess"
//
// Wing Commando's two keywords; prowess is the shared trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "863a5ddb-2d66-4a07-bafc-38290776d5c0",
		Name:            "Mistral Singer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
	})
}
