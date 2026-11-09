package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tolsimir, Midnight's Light — Legendary Creature — Elf Scout {2}{G}{W}{W}
// (Murders at Karlov Manor, #4469):
//
//	"Lifelink
//	 When Tolsimir enters, create Voja Fenstalker, a legendary 5/5 green
//	 and white Wolf creature token with trample.
//	 Whenever a Wolf you control attacks, if Tolsimir attacked this
//	 combat, target creature an opponent controls blocks that Wolf this
//	 combat if able."
//
// The second trigger is one per attacking Wolf (CR 603.2, each Wolf's
// attack is its own event), with an intervening "if" (CR 603.4) checked
// as the Wolf attacks and again as the ability resolves: Tolsimir
// "attacked this combat" is this combat phase's attack records
// (AttackersThisCombatForEffect), so it stays true after he has left. A
// Wolf is read through the layered view, so a changeling counts. The
// blocker is told to block that Wolf (the trigger's event object), the
// same requirement Provoke registers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0e9c7380-7584-43e6-8c31-5d1bb1685fe0",
		Name:            "Tolsimir, Midnight's Light",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tolsimir — create Voja Fenstalker, a legendary 5/5 green and white Wolf creature token with trample",
				Do(CreateToken{Template: TokenCard("5/5 green and white Legendary Wolf named Voja Fenstalker"), N: 1})),
			Targeting(InterveningIf(CondBoard{
				Name: "Tolsimir attacked this combat",
				Test: func(ctx *Context) bool {
					ref, ok := ctx.SourceRef()
					return ok && attackedThisCombat(ctx.Game, ref)
				},
			}, On(game.EventAttack, tolsimirWolfAttacked,
				"Tolsimir — target creature an opponent controls blocks that Wolf this combat if able",
				TargetBlocksTheAttacker)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

// attackedThisCombat is "[object] attacked this combat": declared as an
// attacker in the combat phase in progress (a record, so it holds for
// an object that has since left the battlefield).
func attackedThisCombat(g *game.Game, ref game.ObjectRef) bool {
	for _, a := range g.AttackersThisCombatForEffect() {
		if a == ref {
			return true
		}
	}
	return false
}

// tolsimirWolfAttacked is "whenever a Wolf you control attacks, if
// Tolsimir attacked this combat": the attacker is a Wolf under the
// source's controller and the source is among this combat's attackers.
func tolsimirWolfAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	wolf, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || wolf.Controller != source.Controller || !wolf.HasSubtype("Wolf") {
		return false
	}
	return attackedThisCombat(g, game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch})
}
