package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peter Parker // Amazing Spider-Man — a modal double-faced card whose
// front face transforms (Marvel's Spider-Man; Gwen Stacy // Ghost-Spider
// is the layout precedent).
//
//	Peter Parker                       {1}{W}   0/1
//	Legendary Creature — Human Scientist Hero
//	  When Peter Parker enters, create a 2/1 green Spider creature
//	  token with reach.
//	  {1}{G}{W}{U}: Transform Peter Parker. Activate only as a sorcery.
//
//	Amazing Spider-Man                 {1}{G}{W}{U}   4/4
//	Legendary Creature — Spider Human Hero
//	  Vigilance, reach
//	  Each legendary spell you cast that's one or more colors has
//	  web-slinging {G}{W}{U}.
//
// The back face is registered under this oracle ID plus "#1"
// (game.CatalogKey). The transform is TransformThis, as Gwen Stacy's.
//
// The grant is a GrantedAlternativeCost (ADR 0118 §3) carrying
// web-slinging's two components: the mana and the return of a tapped
// creature you control to its owner's hand (CR 702.188a). "Legendary
// ... that's one or more colors" is PermissionFilter.LegendaryOnly plus
// ColorsAny of all five colours, read off the face being cast. The
// source must be on the battlefield for the static to grant it, so the
// back face never grants web-slinging to itself as a spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     peterParkerOracle,
		Name:         "Peter Parker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Peter Parker — create a 2/1 green Spider creature token with reach",
				Do(CreateToken{Template: TokenCard("2/1 green Spider with reach"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}{G}{W}{U}: Transform Peter Parker. Activate only as a sorcery.",
			Cost:         ManaCost("{1}{G}{W}{U}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})

	Register(Spec{
		OracleID:        peterParkerOracle + "#1",
		Name:            "Amazing Spider-Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "reach"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{
			amazingSpiderManWebSlinging(),
		},
	})
}

// peterParkerOracle is shared with the back face.
const peterParkerOracle = "f1b7488c-c675-42e9-818a-24a50347ec2c"

// amazingSpiderManWebSlinging is "Each legendary spell you cast that's
// one or more colors has web-slinging {G}{W}{U}."
func amazingSpiderManWebSlinging() game.GrantedAlternativeCost {
	gr := WebSlingingForSpellsYouCast("{G}{W}{U}", game.PermissionFilter{LegendaryOnly: true, ColorsAny: "WUBRG"})
	gr.SpellsText = "legendary spell that's one or more colors"
	return gr
}
