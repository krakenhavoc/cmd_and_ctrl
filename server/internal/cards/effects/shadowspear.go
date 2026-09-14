package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadowspear — Legendary Artifact — Equipment for {1} (EDHREC rank
// 314, the highest-ranked Equipment in this batch):
//
//	"Equipped creature gets +1/+1 and has trample and lifelink.
//	 {1}: Permanents your opponents control lose hexproof and
//	 indestructible until end of turn.
//	 Equip {2}"
//
// One mana for a stat line good enough to justify the card, and then
// an ability that is the format's cleanest answer to the two keywords
// that otherwise blank removal. It is played in decks that own no
// other Equipment purely for the second line.
//
// THE SECOND LINE IS A TURN-SCOPED LAYER 6 STATIC THAT TAKES THINGS
// AWAY — StaticUntilEOT (the #279 escape hatch) wrapping the same
// shape RemoveFromAttached has, with the attachment predicate swapped
// for "permanents your opponents control". Three notes on it:
//
//   - The affected SET is re-evaluated on every recompute, not
//     snapshotted, which is what the printed wording means: a
//     permanent that changes control mid-turn changes sides.
//   - "Your opponents" is fixed at ACTIVATION (CR 611.2), because the
//     turn-scoped registration clones the source card as it was; the
//     effect keeps working if the Shadowspear itself is destroyed in
//     response, which is also correct.
//   - It costs {1} and is not a one-shot: activate it once and every
//     removal spell you cast for the rest of the turn connects.
//
// ONE SIMPLIFICATION, strictly weaker: a permanent that ENTERS the
// battlefield after the ability resolves keeps its hexproof and
// indestructible. This engine models a card's printed keywords as a
// layer-6 static stamped with that card's own battlefield timestamp,
// so a permanent arriving later sorts AFTER this removal and its
// grant survives it — the same CR 613.7 ordering Colossus Hammer's
// "loses flying" declares. Everything already on the battlefield when
// you pay the {1} is stripped correctly, which is the case the card
// is activated for.
func init() {
	Register(Spec{
		OracleID:     "8b27326f-e7b8-4a4d-b589-df459246d19a",
		Name:         "Shadowspear",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Permanents that enter after the ability resolves keep their hexproof and indestructible."},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("trample", "lifelink"),
		},
		Activated: []ActivatedAbility{
			{
				Label: "{1}: Permanents your opponents control lose hexproof and indestructible until end of turn",
				Cost:  ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return StaticUntilEOT{
						Label: "Shadowspear — opponents' permanents lose hexproof and indestructible",
						Ability: game.StaticAbility{
							Layer: game.Layer6Ability,
							AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
								return target.Controller != source.Controller
							},
							Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
								kept := c.Abilities[:0]
								for _, a := range c.Abilities {
									if a == "hexproof" || a == "indestructible" {
										continue
									}
									kept = append(kept, a)
								}
								c.Abilities = kept
							},
						},
					}.Apply(NewContext(g, item))
				},
			},
			EquipAbility("{2}"),
		},
	})
}
