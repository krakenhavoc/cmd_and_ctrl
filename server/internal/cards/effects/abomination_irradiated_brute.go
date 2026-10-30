package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abomination, Irradiated Brute — Legendary Creature — Gamma Villain
// {2}{G}{G}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 Whenever a Gamma or Villain you control deals combat damage to a
//	 player, put that many +1/+1 counters on Abomination."
//
// Seafloor Oracle's per-creature combat-damage condition
// (combatDamageToPlayerBy) narrowed to two creature types, read off the
// damage source; each Gamma or Villain that connects is its own trigger
// (the card says "a", not "one or more"), counting its own damage. The
// types are words of the card's text (CR 612.2). The counters go on
// Abomination only while he is still the permanent that triggered
// (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7436d37-c331-4524-b5d4-861e9d25bfc5",
		Name:            "Abomination, Irradiated Brute",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, msc16GammaOrVillainDealtCombatDamageToAPlayer,
				"Abomination, Irradiated Brute — put that many +1/+1 counters on him",
				Do(Instruction(func(ctx *Context) error {
					n := ctx.Trigger().Event.Amount
					if n <= 0 || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					return CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}))),
		},
	})
}

// msc16GammaOrVillainDealtCombatDamageToAPlayer is "whenever a Gamma or
// Villain you control deals combat damage to a player": a creature its
// controller controls dealt combat damage to a player, and it has either
// type as it is now (a damage source already gone does not fire, which is
// weaker than printed, never stronger).
func msc16GammaOrVillainDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && (src.HasSubtype(source.TextSubtype("Gamma")) || src.HasSubtype(source.TextSubtype("Villain")))
}
