package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storm, Queen of Wakanda — Legendary Creature — Mutant Noble Hero
// {3}{G}{W}, 4/5 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Storm attacks, until end of turn, another target attacking
//	 creature gains flying and gets +X/+0, where X is Storm's power.
//	 Whenever a creature with flying attacks you, Storm deals damage equal
//	 to her power to that creature."
//
// Both numbers are Storm's power as the ability resolves, and the last
// known power once she has left (sourcePowerNow, CR 608.2h), so a Storm
// that was destroyed in response still gives +X/+0 and still deals her
// damage. The first trigger targets another creature that is attacking
// (Storm, the source, is excluded), and both halves lock to that creature.
// The second watches an attack declaration aimed at the player (not at a
// planeswalker or battle: "attacks you"), by a creature of another player
// that has flying as it is declared, read through the layered abilities.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8e585c5-1de3-4558-8a2f-0d9fb07edad6",
		Name:            "Storm, Queen of Wakanda",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(WheneverThisAttacks(stormGrantLabel, stormGrantAttacker),
					Another(TargetCreature("another target attacking creature", AttackingCreature()))),
				ForTargets(BenefitUntilEOT(0, 0, 0, "flying"))),
			On(game.EventAttack, stormFlyerAttacksYou, stormDamageLabel, stormDamageFlyer),
		},
	})
}

const (
	stormGrantLabel  = "Storm, Queen of Wakanda — another target attacking creature gains flying and gets +X/+0, where X is Storm's power"
	stormDamageLabel = "Storm, Queen of Wakanda — Storm deals damage equal to her power to that creature"
)

func stormGrantAttacker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	x := sourcePowerNow(ctx)
	if err := (BoostUntilEOT{Target: id, Power: x, Label: "Storm, Queen of Wakanda — +X/+0"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Storm, Queen of Wakanda — flying"}.Apply(ctx)
}

// stormFlyerAttacksYou is the When of the second ability: a creature with
// flying that another player controls was declared attacking Storm's
// controller.
func stormFlyerAttacksYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.Target != source.Controller || ev.Actor == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && game.HasKeyword(&c, "flying")
}

func stormDamageFlyer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	attacker := item.Trigger.Event.CardID
	if !onBattlefield(g, attacker) {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: attacker, Amount: sourcePowerNow(ctx)}.Apply(ctx)
}
