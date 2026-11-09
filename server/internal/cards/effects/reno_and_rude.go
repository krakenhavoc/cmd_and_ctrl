package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reno and Rude — Legendary Creature — Human Assassin {1}{B}, 2/1 (Final
// Fantasy, #4462):
//
//	"Menace
//	 Whenever Reno and Rude deals combat damage to a player, exile the top
//	 card of that player's library. Then you may sacrifice another
//	 creature or artifact. If you do, you may play the exiled card this
//	 turn, and mana of any type can be spent to cast it."
//
// The exile is plain (no permission yet); its tail asks the optional
// sacrifice (a resolution-time pick of zero or one, so a hexproof
// permanent of yours can go), and only a real sacrifice grants the play
// permission over the card that landed in exile: a this-turn "play" (so a
// land counts, with its land drop and timing, and a spell its own timing),
// with any type of mana spendable on it (AnyType, CR 106.6). An empty
// library exiles nothing and asks nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f749a6d6-2c55-439f-bb02-e95fa54c2ef3",
		Name:            "Reno and Rude",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Reno and Rude — exile the top card of that player's library; you may sacrifice another creature or artifact to play it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					victim := item.Trigger.Event.Target
					if g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					top := g.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), victim, 1)
					if len(top) == 0 {
						return nil
					}
					return g.ExileCardKeyedForEffect(top[0], ctx.ItemTail(renoAndRudeExiled))
				}),
		},
	})
}

// renoAndRudeExiled is the exile landed: ask the optional sacrifice, with
// the exiled card as the pick's carry. An on-disk key.
var renoAndRudeExiled = CardTail("reno-and-rude/exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	self := ctx.Source()
	return ChoosePermanents{
		Question: "Reno and Rude — you may sacrifice another creature or artifact to play the exiled card",
		Keyed:    renoAndRudeSacrificed,
		Carry:    o.IDs,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.InstanceID != self && (c.IsCreature() || c.IsArtifact()) {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 1
		},
	}.Apply(ctx)
})

// renoAndRudeSacrificed is "if you do, you may play the exiled card this
// turn, and mana of any type can be spent to cast it", keyed (#4290).
var renoAndRudeSacrificed = RegisterPermanentsPick("prompt-run/reno-and-rude-sacrifice", PermanentsDef{
	Sacrifice: true,
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		if picked.Count() == 0 {
			return nil
		}
		g := ctx.Game
		var cards []game.Card
		for _, id := range r.Carry {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
				continue
			}
			if c, ok := g.LookupCardForEffect(id); ok {
				cards = append(cards, c)
			}
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     ctx.Controller(),
			Zone:       game.ZoneExile,
			Duration:   g.UntilEndOfTurnDuration(),
			AnyColor:   true,
			AnyType:    true,
			Source:     ctx.Source(),
			SourceName: "Reno and Rude",
			Label:      "Reno and Rude — play the exiled card this turn, spending mana of any type",
		}, cards)
		return nil
	},
})
