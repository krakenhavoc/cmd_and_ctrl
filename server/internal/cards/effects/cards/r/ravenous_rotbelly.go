package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ravenous Rotbelly — Creature — Zombie Horror {4}{B}, 4/5 (Midnight
// Hunt Commander, #5497):
//
//	"When this creature enters, you may sacrifice up to three Zombies.
//	 When you sacrifice one or more Zombies this way, each opponent
//	 sacrifices that many creatures of their choice."
//
// The first sentence is God-Eternal Bontu's pick at resolution, with a
// ceiling of three and a floor of zero (the "may" and the "up to"), over
// the Zombies you control through effective subtypes (a changeling
// counts) and the Rotbelly itself, a Zombie, among them. The chosen
// Zombies are sacrificed together. The second is a CR 603.12 reflexive
// trigger, made only when at least one Zombie was sacrificed, carrying the
// number sacrificed (Augusta's "that many"). It is Portal to Phyrexia's
// edict with the count: each opponent in APNAP order names that many
// creatures, one choice each, and a player with fewer sacrifices what
// they have (CR 701.21a); nothing is targeted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f0738ce-1f07-422c-9895-8ac3c65f9054",
		Name:         "Ravenous Rotbelly",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisEnters("Ravenous Rotbelly — you may sacrifice up to three Zombies; each opponent sacrifices that many creatures", ravenousRotbellyEnters)),
		},
	})
}

func ravenousRotbellyEnters(g *game.Game, item *game.StackItem) error {
	zombie := g.TextReaderForEffect(item).TextSubtype("Zombie")
	return effects.ChoosePermanents{
		Question: "Ravenous Rotbelly — sacrifice up to three Zombies",
		Keyed:    ravenousRotbellySacrificed,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.HasSubtype(zombie) {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 3
		},
	}.Apply(effects.NewContext(g, item))
}

var ravenousRotbellySacrificed = effects.RegisterPermanentsPick("prompt-run/ravenous-rotbelly-sacrifice", effects.PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		n := picked.Count()
		if n == 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label:  "Ravenous Rotbelly — each opponent sacrifices that many creatures of their choice",
			Body:   ravenousRotbellyEdictBody,
			Params: game.EffectParams{Amount: n},
		}.Apply(ctx)
	},
})

var ravenousRotbellyEdictBody = game.ReflexiveBody("ravenous-rotbelly/edict", effects.SimpleBody(ravenousRotbellyEdict), nil)

// ravenousRotbellyEdict asks each opponent, APNAP from the active player,
// for that many creatures.
func ravenousRotbellyEdict(g *game.Game, item *game.StackItem) error {
	var asks []uuid.UUID
	if n := len(g.Seats); n > 0 {
		start := g.Turn.ActiveSeat
		for i := 0; i < n; i++ {
			p := g.Seats[(start+i)%n]
			if p == nil || p.Eliminated || p.ID == item.Controller {
				continue
			}
			for k := 0; k < item.Params.Amount; k++ {
				asks = append(asks, p.ID)
			}
		}
	}
	g.PlayersSacrificeForEffect(item.SourceCardID, asks,
		effects.SacrificeSpec("a creature", effects.Creature()), "Sacrifice a creature")
	return nil
}
