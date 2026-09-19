package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oona, Queen of the Fae — Legendary Creature — Faerie Wizard
// {3}{U/B}{U/B}{U/B}, 5/5:
//
//	"Flying
//	 {X}{U/B}: Choose a color. Target opponent exiles the top X cards
//	 of their library. For each card of the chosen color exiled this
//	 way, create a 1/1 blue and black Faerie Rogue creature token with
//	 flying."
//
// Flying is a printed keyword the importer stamps. The ability is an
// X activated ability with a hybrid mana symbol and a target opponent;
// as it resolves, #742's resolution-time prompt asks its controller
// for a colour, and the continuation does the rest in printed order:
// exile the top X (an exile, not a mill — MillToZone to exile), then
// one Faerie Rogue per exiled card of the chosen colour. A
// multicoloured card counts if the chosen colour is one of its
// colours; lands and other colourless cards never count.
//
// The colour is chosen before any card is exiled, as printed, so the
// choice cannot be made knowing what was exiled.
//
// #893: "exiled this way" is the cards that ARRIVED in exile
// (CR 400.7), so the Faeries come from the continuation rather than
// from the line after the exile. An opponent whose commander comes off
// the top of their library is asked whether it goes to the command
// zone instead (CR 903.9); until they answer, nothing has moved, and
// if they accept, that card was not exiled this way and pays for no
// Faerie. Both halves used to be wrong in the same direction — one
// Faerie too many, handed over before the question was answered.
//
// No simplification.
func init() {
	const label = "{X}{U/B}: Choose a color. Target opponent exiles the top X cards of their library. " +
		"For each card of the chosen color exiled this way, create a 1/1 blue and black Faerie Rogue creature token with flying."
	Register(Spec{
		OracleID:     "6052822d-47a2-4d69-a32d-40cdd600d7a9",
		Name:         "Oona, Queen of the Fae",
		XMatters:     true,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   label,
			Cost:    ManaCost("{X}{U/B}"),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var victim uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						victim = t.ID
					}
				}
				if victim == uuid.Nil {
					return nil
				}
				x := ctx.X()
				ChooseColorThen(game.ColorForFilter, g, item.Controller, item.SourceCardID, "Oona, Queen of the Fae — choose a color",
					func(g *game.Game, color string) error {
						return oonaExileAndMakeFaeries(NewContext(g, item), victim, x, color)
					})
				return nil
			},
		}},
	})
}

func oonaExileAndMakeFaeries(ctx *Context, victim uuid.UUID, x int, color string) error {
	return MillToZone{
		Player: victim,
		N:      x,
		To:     game.ZoneExile,
		Then: func(ctx *Context, exiled []uuid.UUID) error {
			n := 0
			for _, id := range exiled {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.HasColor(color) {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("1/1 blue and black Faerie Rogue with flying"),
				N:          n,
			}.Apply(ctx)
		},
	}.Apply(ctx)
}
