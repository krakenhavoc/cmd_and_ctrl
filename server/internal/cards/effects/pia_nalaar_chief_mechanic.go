package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pia Nalaar, Chief Mechanic — Legendary Creature — Human Artificer {G}{U}{R},
// 2/4:
//
//	"Whenever one or more artifact creatures you control deal combat damage
//	 to a player, you get {E}{E} (two energy counters).
//	 At the beginning of your end step, you may pay one or more {E}. If you
//	 do, create an X/X colorless Vehicle artifact token named Nalaar
//	 Aetherjet with flying and crew 2, where X is the amount of {E} paid this
//	 way."
//
// The first ability collapses per damaged player (CR 603.2c), the way a
// "one or more" combat damage trigger does. The second is a pay-any-amount
// prompt with a floor of one: the payer picks X, and the token is made at
// that size. The Aetherjet is a printed Vehicle token (crew 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31cb87bf-1020-4fd5-95f7-8ecb197d823b",
		Name:         "Pia Nalaar, Chief Mechanic",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && dealer.Controller == source.Controller && dealer.IsCreature() && dealer.IsArtifact()
			}, "Pia Nalaar, Chief Mechanic — you get {E}{E}", Do(GetEnergy{N: 2}))),
			AtYourEndStep("Pia Nalaar, Chief Mechanic — you may pay one or more {E}; if you do, create an X/X Vehicle token named Nalaar Aetherjet",
				func(g *game.Game, item *game.StackItem) error {
					return PayEnergyAmount{
						Min:      1,
						Question: "Pia Nalaar, Chief Mechanic — pay one or more {E} for an X/X Nalaar Aetherjet",
						Unit:     game.PayAmountPower,
						Goal:     AsMuchAsYouCan,
						Then:     piaNalaarAetherjet,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// piaNalaarAetherjet is the paid half: an X/X Nalaar Aetherjet, X the {E}
// paid. An on-disk key: never renamed, never reused.
var piaNalaarAetherjet = NumberThen("number/pia-nalaar-create-nalaar-aetherjet",
	func(ctx *Context, r game.NumberAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return CreateToken{Controller: ctx.Controller(), Template: CmdrB23NalaarAetherjetToken(r.Number), N: 1}.Apply(ctx)
	})
