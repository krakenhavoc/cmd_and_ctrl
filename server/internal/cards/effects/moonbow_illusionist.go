package effects

// Moonbow Illusionist — Creature — Moonfolk Wizard {2}{U}, 2/1:
//
//	"Flying
//	 {2}, Return a land you control to its owner's hand: Target land becomes the basic land type of your choice until end of turn."
//
// ADR 0109 §1 (#1881): CR 305.7 from a resolved ability. The basic land
// type is chosen as the ability resolves (CR 608.2), by its controller,
// through effects.ChooseBasicLandTypeThen; a land that has become an
// illegal target by then is left alone and nothing is asked (CR 608.2b).
// Until end of turn the land's land types are replaced by the chosen one
// (its other subtypes stay, CR 205.1a), it loses the abilities its rules
// text gives it, keeps any another effect granted it, and taps for the
// chosen type's colour (CR 305.6).
//
// The return is a cost (CR 602.2b), named at activation; the land
// returned may be the one targeted, which then is an illegal target and
// nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "411137a5-7054-42bc-a731-88a15e5e6b9d",
		Name:            "Moonbow Illusionist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Return a land you control to its owner's hand: Target land becomes the basic land type of your choice until end of turn.",
			Cost:    Plus(ManaCost("{2}"), ReturnAPermanentToHand("a land you control", Land())),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesUntilEOT("Moonbow Illusionist"),
		}},
	})
}
