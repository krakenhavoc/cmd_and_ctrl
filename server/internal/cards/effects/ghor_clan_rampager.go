package effects

// Ghor-Clan Rampager — Creature — Beast {2}{R}{G}, 4/4 (Gatecrash, #5565):
//
//	"Trample
//	 Bloodrush — {R}{G}, Discard this card: Target attacking creature gets
//	 +4/+4 and gains trample until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b). Trample is the engine keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ec7a9e7f-742b-4344-983e-73372c6a1760",
		Name:            "Ghor-Clan Rampager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []ActivatedAbility{gtcBloodrush("Ghor-Clan Rampager", "{R}{G}", 4, 4, "trample")},
	})
}
