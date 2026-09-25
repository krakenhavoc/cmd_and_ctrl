package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza, Lord High Artificer — Legendary Creature — Human Artificer
// {2}{U}{U}, 1/4:
//
//	"When Urza enters, create a 0/0 colorless Construct artifact
//	 creature token with "This token gets +1/+1 for each artifact you
//	 control."
//	 Tap an untapped artifact you control: Add {U}.
//	 {5}: Shuffle your library, then exile the top card. Until end of
//	 turn, you may play that card without paying its mana cost."
//
// The Construct is a real ADR 0083 token template with its own
// static ability (urzaConstructToken below) rather than the
// Simulacrum Synthesizer workaround (a static on the GRANTOR that
// sizes its own Constructs and stops working the instant it leaves):
// the self-scaling P/T lives on the token itself, so it keeps sizing
// after Urza dies, exactly as a printed Construct does.
//
// "Tap an untapped artifact you control" is a mana ability with NO
// {T} on Urza — Springleaf Drum's TapOthers shape, filtered to
// artifacts, and Urza taps no artifact of its own to pay it (it isn't
// one).
//
// The {5} ability shuffles, then exiles the new top card with a
// standing "play it free, until end of turn" permission — Etali,
// Primal Storm's Cost: "{0}" grant with a Duration added, and
// CastOnly left false because the printed word is "play", not "cast"
// (a land off the top is playable, exactly as the card says).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e87906d2-db1a-4e19-b910-adb4eb339945",
		Name:         "Urza, Lord High Artificer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Urza, Lord High Artificer — create a Construct",
				Do(CreateToken{Template: UrzaConstructToken(), N: 1})),
		},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{TapOthers: &game.TapOthersCost{
				Count:  1,
				Filter: TargetPermanent("an untapped artifact you control", Artifact()),
				Label:  "an untapped artifact you control",
			}},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{5}: Shuffle your library, then exile the top card. Until end of turn, you may play that card without paying its mana cost.",
			Cost:  ManaCost("{5}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (ShuffleLibrary{Player: item.Controller}).Apply(ctx); err != nil {
					return err
				}
				_, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 1, game.CastPermission{
					Cost:     "{0}",
					Duration: g.UntilEndOfTurnDuration(),
				})
				return err
			},
		}},
	})
}

// UrzaConstructToken is the 0/0 Construct Urza's ETB creates.
func UrzaConstructToken() game.Card { return tokenFromCatalog(printedUrzaConstructToken) }

// printedUrzaConstructToken carries its own printed ability: "This
// token gets +1/+1 for each artifact you control." A self-only
// Layer 7c modify (not a CDA — the base 0/0 is printed, the bonus is
// a boost on top of it), reading the token's OWN controller so a
// changed-control Construct sizes off its new controller's board.
func printedUrzaConstructToken() tokenTemplate {
	return tokenTemplate{
		Slug: "urza-construct",
		Card: game.Card{
			Name:      "Construct",
			TypeLine:  "Token Artifact Creature — Construct",
			Power:     0,
			Toughness: 0,
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, cand := range g.BattlefieldCardsForEffect() {
					if cand.Controller == target.Controller && cand.IsArtifact() {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
		Text: "This token gets +1/+1 for each artifact you control.",
	}
}
