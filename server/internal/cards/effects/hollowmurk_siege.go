package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hollowmurk Siege — Enchantment {B}{G}:
//
//	"As this enchantment enters, choose Sultai or Abzan.
//	 • Sultai — Whenever a counter is put on a creature you control,
//	   draw a card. This ability triggers only once each turn.
//	 • Abzan — Whenever you attack, put a +1/+1 counter on target
//	   attacking creature. It gains menace until end of turn."
//
// #1647: the third of the three remaining Sieges.
//
//   - Sultai watches EventCounterPlaced for ANY counter kind, unlike
//     Exemplar of Light's own-instance +1/+1 watcher — it has to read
//     the placed-on card off ev.Target rather than trust the source,
//     and it has to tell a placement from a removal generically
//     (hollowmurkCounterWasPlaced), because the printed clause names
//     no kind. "Once each turn" is Exemplar's own
//     b11TriggeredThisTurn, keyed off this ability's label.
//   - Abzan is Gimli's Reckless Might's shape one clause simpler:
//     "whenever you attack" is ONE trigger per attack declaration
//     (OncePerBatch), not one per attacker, with a target clause on
//     the trigger itself rather than a fresh TargetsFrom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31061e34-042e-40c3-99ab-752795ab4324",
		Name:         "Hollowmurk Siege",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Hollowmurk Siege", "Sultai", "Abzan"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Sultai", On(game.EventCounterPlaced, hollowmurkSultaiApplies,
				b1647HollowmurkSultaiLabel, Do(DrawCards{N: 1}))),
			WhenChosen("Abzan", OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Hollowmurk Siege — put a +1/+1 counter on target attacking creature", hollowmurkAbzanBody),
				TargetCreature("target attacking creature", AttackingCreature())))),
		},
	})
}

const b1647HollowmurkSultaiLabel = "Hollowmurk Siege — draw a card"

// hollowmurkSultaiApplies is "a counter is put on a creature you
// control", any kind, once each turn.
func hollowmurkSultaiApplies(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCounterPlaced {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !target.IsCreature() || target.Controller != source.Controller {
		return false
	}
	if !hollowmurkCounterWasPlaced(ev, g) {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, b1647HollowmurkSultaiLabel)
}

// hollowmurkCounterWasPlaced is b11CountersWerePlaced without the kind
// parameter: EventCounterPlaced fires for a removal too and carries
// only the post-change total for its OWN kind, so "a counter is put"
// (no kind named) still has to walk the log back to the same
// (target, kind) pair to tell a placement from a removal.
func hollowmurkCounterWasPlaced(ev game.Event, g *game.Game) bool {
	before := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		prev := g.Events[i]
		if prev.Seq >= ev.Seq {
			continue
		}
		if (prev.Kind == game.EventETB || prev.Kind == game.EventTokenCreated) && prev.CardID == ev.Target {
			break
		}
		if prev.Kind == game.EventCounterPlaced && prev.Target == ev.Target && prev.Label == ev.Label {
			before = prev.Amount
			break
		}
	}
	return ev.Amount > before
}

// hollowmurkAbzanBody is the Abzan body.
func hollowmurkAbzanBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	t := targets[0]
	if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   t.ID,
		Keywords: []string{"menace"},
		Label:    "Hollowmurk Siege — menace until end of turn",
	}.Apply(ctx)
}
