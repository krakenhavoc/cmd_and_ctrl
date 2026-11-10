package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kylox, Visionary Inventor — Legendary Creature — Lizard Artificer
// {5}{U}{R}, 4/4 (Murders at Karlov Manor, #4469):
//
//	"Menace, ward {2}, haste
//	 Whenever Kylox attacks, sacrifice any number of other creatures,
//	 then exile the top X cards of your library, where X is their total
//	 power. You may cast any number of instant and/or sorcery spells from
//	 among the exiled cards without paying their mana costs."
//
// The sacrifice is God-Eternal Bontu's "any number of other" pick (zero
// allowed, all sacrificed at once). X is the total power the creatures
// last had on the battlefield (CR 608.2h, the ruling). The exile and the
// free casts are ExileTopAnyNumberFreeCast: cast during the resolution
// (CR 608.2g), instants and sorceries only, timing ignored, X in a cost
// is 0, and what isn't cast stays in exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "37f6bbb8-2136-4f03-966c-85d72180e71f",
		Name:            "Kylox, Visionary Inventor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "haste"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Kylox, Visionary Inventor — ward {2}"),
			WheneverThisAttacks("Kylox, Visionary Inventor — sacrifice any number of other creatures, then exile that much power's worth of cards and cast spells from among them",
				kyloxAttack),
		},
	})
}

func kyloxAttack(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	return ChoosePermanents{
		Question: "Kylox, Visionary Inventor — sacrifice any number of other creatures",
		Keyed:    kyloxSacrificed,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.InstanceID != self && c.IsCreature() {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 0
		},
	}.Apply(NewContext(g, item))
}

var kyloxSacrificed = RegisterPermanentsPick("prompt-run/kylox-sacrifice", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		x := 0
		for _, id := range picked.Cards() {
			x += b17LastKnownPowerOffBattlefield(ctx.Game, id)
		}
		return ExileTopAnyNumberFreeCast(ctx, x, game.PermissionFilter{InstantOrSorceryOnly: true})
	},
})
