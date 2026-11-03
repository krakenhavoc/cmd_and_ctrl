package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inspired Skypainter // Maestro's Gift — Creature — Lizard Wizard
// {U}{R}, 2/2, with the prepare spell Maestro's Gift, Sorcery {3}{U}{R}:
//
//	"Flying
//	 When this creature enters and whenever one or more creature tokens
//	 you control deal combat damage to a player, this creature becomes
//	 prepared. (While it's prepared, you may cast a copy of its spell.
//	 Doing so unprepares it.)"
//
//	Maestro's Gift: "Create a token that's a copy of target creature you
//	 control. That token gains haste until end of turn."
//
// A preparation card (CR 722): the permanent under the bare oracle ID and
// its prepare spell under "#1". The engine makes the copy in exile as the
// creature becomes prepared and unprepares it as the copy is cast; a
// creature already prepared does not become prepared again. The trigger
// is one printed ability with two conditions: the Skypainter entering,
// and one or more of its controller's creature tokens dealing combat
// damage to a player, once per player they hit in a damage step
// (CR 603.2c).
//
// Maestro's Gift copies the creature's copiable values (CR 707.2), so the
// token has none of its counters, Auras or other effects; the haste is
// not part of the copy, it is an effect on the new token that ends at
// cleanup.
//
// No simplification.
func init() {
	const oracle = "d751580b-bb72-4d1d-be27-d0410f883b2a"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Inspired Skypainter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(OnAny([]game.EventKind{game.EventETB, game.EventDealDamage},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventETB {
						return Self(ev, source, lki, g)
					}
					return socB3CreatureTokenYouControlDealtCombatDamageToAPlayer(ev, source, lki, g)
				}, "Inspired Skypainter — this creature becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				})),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Maestro's Gift",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				copyOf := FirstLegalBattlefieldTarget(ctx)
				if copyOf == uuid.Nil {
					return nil
				}
				ctx.RecordStep(maestrosGiftRecord, nil, int(b25LastEventSeq(ctx.Game)))
				return CreateTokenCopy{Controller: ctx.Controller(), Copy: copyOf, N: 1}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				r := ctx.StepRecord(maestrosGiftRecord)
				if len(r.Ns) == 0 {
					return nil
				}
				for _, id := range b27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(r.N())) {
					if err := (ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddKeywordsMod("haste")},
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Maestro's Gift — the token gains haste until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}

const maestrosGiftRecord = "maestros-gift/cursor"
