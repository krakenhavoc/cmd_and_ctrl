package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// parapetThrasherLabel is the trigger's stack label, and with it the
// key its "hasn't been chosen this turn" memory is kept under.
const parapetThrasherLabel = "Parapet Thrasher — Dragons dealt combat damage to an opponent"

// Parapet Thrasher — Creature — Dragon {2}{R}{R}, 4/3:
//
//	"Flying
//	 Whenever one or more Dragons you control deal combat damage to an
//	 opponent, choose one that hasn't been chosen this turn —
//	 • Destroy target artifact that opponent controls.
//	 • This creature deals 4 damage to each other opponent.
//	 • Exile the top card of your library. You may play it this turn."
//
// "One or more … to an opponent" is one trigger per opponent the
// Dragons connect with in a damage step (OncePerBatchPerPlayer, CR
// 603.2c), and "that opponent" is the damaged player the trigger
// carries (item.Trigger.Event.Target). ChooseOneNotChosenThisTurn
// (ADR 0097) spreads a turn's triggers over the three bullets: two
// opponents hit at once are two triggers taking two different bullets,
// and a first-strike step and a regular one are two batches, so a
// Dragon with double strike can use a second bullet in the same combat.
//
// "Each other opponent" is every opponent of the controller except the
// damaged one, and the damage comes from this creature.
//
// Simplification, declared: a mode's target clause is fixed on the
// card and is never handed the trigger, so the artifact bullet offers
// any artifact an OPPONENT controls and does nothing at resolution
// unless THAT opponent — the one who was dealt damage — controls it.
// The bullet can therefore be chosen and wasted where the printed card
// would not offer it; it never destroys an artifact the printed card
// could not.
func init() {
	Register(Spec{
		OracleID:        "722e23d5-4f2a-43cd-bbd5-baf56b76e68d",
		Name:            "Parapet Thrasher",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The artifact choice offers any opponent's artifacts, and only destroys one controlled by the opponent who was dealt damage."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			parapetThrasherTrigger(),
		},
	})
}

func parapetThrasherTrigger() game.TriggeredAbility {
	t := WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(OfCreatureType("Dragon"),
		parapetThrasherLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosenThisTurn(
		ModeDoing("Destroy target artifact that opponent controls.",
			TargetPermanent("target artifact that opponent controls", Artifact(), OpponentControls()),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || c.Controller != item.Trigger.Event.Target {
					return nil
				}
				return DestroyTarget{Target: t.ID}.Apply(ctx)
			}),
		ModeDoing("This creature deals 4 damage to each other opponent.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				that := item.Trigger.Event.Target
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, opp := range ctx.Opponents() {
						if opp == that {
							continue
						}
						if err := ctx.Game.DealDamageToPlayerForEffect(item.SourceCardID, opp, 4); err != nil {
							return err
						}
					}
					return nil
				})
			}),
		ModeDoing("Exile the top card of your library. You may play it this turn.", nil,
			exileTopCardYouMayPlayThisTurn),
	)
	return t
}
