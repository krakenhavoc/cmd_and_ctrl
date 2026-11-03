package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kirol, History Buff // Pack a Punch — a preparation card (CR 722, ADR 0090),
// Secrets of Strixhaven (#4456):
//
//	Kirol, History Buff {R}{W}, Legendary Creature — Vampire Cleric 2/3:
//	"Whenever one or more cards leave your graveyard, Kirol becomes
//	 prepared. (While it's prepared, you may cast a copy of its spell.
//	 Doing so unprepares it.)"
//	Pack a Punch {1}{R}{W}, Sorcery:
//	"Mill a card. Put two +1/+1 counters on target creature. It gains trample
//	 until end of turn."
//
// The trigger is Spirit of Resilience's: one per batch of cards leaving the
// graveyard together (CR 603.2c), a cast from the graveyard included
// (whenCardsLeaveYourGraveyard). Becoming prepared does nothing for a
// creature that is already prepared (CR 722.3a).
//
// Pack a Punch mills first, then puts the counters on the target and gives
// it trample, in printed order; the counters and the trample go to the
// target only if it is still legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	const id = "df1c0d06-2109-4297-9271-8a003fc892bc"
	Register(Spec{
		OracleID:     id,
		Name:         "Kirol, History Buff",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(leftGraveyardEvents, whenCardsLeaveYourGraveyard,
				"Kirol, History Buff — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				})),
		},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Pack a Punch",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "trample"))},
		OnResolve:    ResolveSteps(MillCards{N: 1}, Instruction(packAPunchCountersAndTrample)),
	})
}

// packAPunchCountersAndTrample is the spell's second and third sentences.
func packAPunchCountersAndTrample(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Pack a Punch — trample until end of turn"}.Apply(ctx)
}
