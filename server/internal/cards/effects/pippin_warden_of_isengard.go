package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pippin, Warden of Isengard — Legendary Creature — Halfling Advisor {B}{G},
// 2/2 (Tales of Middle-earth Commander, #5948):
//
//	"Partner with Merry, Warden of Isengard (When this creature enters, target
//	 player may put Merry into their hand from their library, then shuffle.)
//	 {1}, {T}: Create a Food token.
//	 {T}, Sacrifice four Foods: Other creatures you control get +3/+3 and gain
//	 haste until end of turn. Activate only as a sorcery."
//
// Partner with is the shared entry trigger (PartnerWith, CR 702.124j) and
// the deck-construction rule the validator reads off the text. The Food maker
// is Gilded Goose's. The pump is Overrun's two layers over the other
// creatures you control as the ability resolves (CR 611.2c), the Foods paid
// as the cost (four distinct Foods, the Foods' own sacrifice the way Bosco's
// is); it is sorcery speed, and, with {T} in its cost, subject to summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "421ec8a8-2b20-4296-b8f7-010acc3f8097",
		Name:         "Pippin, Warden of Isengard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Pippin, Warden of Isengard", "Merry, Warden of Isengard"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Create a Food token.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Effect:  Do(CreateToken{Template: FoodToken(), N: 1}),
			},
			{
				Label:        "{T}, Sacrifice four Foods: Other creatures you control get +3/+3 and gain haste until end of turn. Activate only as a sorcery.",
				Cost:         Plus(TapCost(), SacrificeN(4, "four Foods", HasSubtype("Food"))),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					self := item.SourceCardID
					others := func(_ *game.Game, you uuid.UUID, c game.Card) bool {
						return c.IsCreature() && c.Controller == you && c.InstanceID != self
					}
					if err := (BoostUntilEOT{Match: others, Power: 3, Toughness: 3, Label: "Pippin, Warden of Isengard — +3/+3"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: others, Keywords: []string{"haste"}, Label: "Pippin, Warden of Isengard — haste"}.Apply(ctx)
				},
			},
		},
	})
}
