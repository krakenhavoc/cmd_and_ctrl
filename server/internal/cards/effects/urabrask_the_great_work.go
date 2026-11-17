package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urabrask // The Great Work — a transforming Praetor whose back face is a
// Saga (March of the Machine).
//
// Front face, Legendary Creature — Phyrexian Praetor {2}{R}{R}, 4/4:
//
//	"First strike
//	 Whenever you cast an instant or sorcery spell, Urabrask deals 1 damage
//	 to target opponent. Add {R}.
//	 {R}: Exile Urabrask, then return it to the battlefield transformed
//	 under its owner's control. Activate only as a sorcery and only if
//	 you've cast three or more instant and/or sorcery spells this turn."
//
// Back face, Enchantment — Saga:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — This Saga deals 3 damage to target opponent and each creature they
//	     control.
//	 II — Create three Treasure tokens.
//	 III — Until end of turn, you may cast instant and sorcery spells from
//	       any graveyard. If a spell cast this way would be put into a
//	       graveyard, exile it instead. Exile this Saga, then return it to
//	       the battlefield (front face up)."
//
// The flip is ExileAndReturnTransformed (ADR 0079's second verb, as Jin-
// Gitaxias); the back face registers under "<oracle_id>#1". The cast trigger
// is one targeted ability, damage and mana both: if its target is gone as it
// resolves it does nothing, the mana included. The activation reads the cast
// tally (Game.SpellsCastThisTurn), bumped as each spell is cast, so the
// third spell counts the moment it is on the stack.
//
// Chapter I is one damage event from the Saga to the target opponent and to
// every creature that player controls. Chapter III opens every graveyard to
// its controller's instants and sorceries for the turn, one standing
// permission per seat (Shaman's Trance's shape), paid at the card's own cost
// and timing; a spell cast that way is exiled if it would go to a graveyard
// from the stack (ExileInsteadOfGraveyard). The Saga then flickers, returning
// front face up as a new object (CR 712.14a); the permissions are the
// resolution's, not the Saga's, and outlast it.
//
// No simplification.
func init() {
	const castLabel = "Urabrask — 1 damage to target opponent, add {R}"
	Register(Spec{
		OracleID:        urabraskOracleID,
		Name:            "Urabrask",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverYouCast(Or(Instant(), Sorcery()), castLabel, func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t := ctx.LegalTargets()
					if len(t) == 0 || t[0].Kind != game.TargetPlayer {
						return nil
					}
					if err := (DealDamage{Source: item.SourceCardID, Target: t[0].ID, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
					return AddMana{Produced: "{R}"}.Apply(ctx)
				}),
				TargetPlayer("target opponent", Opponent())),
		},
		Activated: []ActivatedAbility{{
			Label: "{R}: Exile Urabrask, then return it to the battlefield transformed under its owner's control. " +
				"Activate only as a sorcery and only if you've cast three or more instant and/or sorcery spells this turn.",
			Cost:         ManaCost("{R}"),
			SorcerySpeed: true,
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return g.SpellsCastThisTurn[controller].InstantSorcery >= 3
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})

	chapterOne := ChapterTrigger(1, SagaChapterLabel("The Great Work", 1, "3 damage to target opponent and each creature they control"),
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			t := ctx.LegalTargets()
			if len(t) == 0 || t[0].Kind != game.TargetPlayer {
				return nil
			}
			opp := t[0].ID
			return g.DamageInstanceForEffect(func() error {
				if err := (DealDamage{Source: item.SourceCardID, Target: opp, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
				return damageEachMatching(ctx, And(Creature(), ControlledBy(opp)), 3)
			})
		})
	chapterOne.Targets = TargetPlayer("target opponent", Opponent())

	Register(Spec{
		OracleID:     urabraskOracleID + "#1",
		Name:         "The Great Work",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(chapterOne, game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}}),
			ChapterTrigger(2, SagaChapterLabel("The Great Work", 2, "create three Treasure tokens"),
				Do(CreateToken{Template: TreasureToken(), N: 3})),
			ChapterTrigger(3, SagaChapterLabel("The Great Work", 3, "you may cast instant and sorcery spells from any graveyard this turn, then exile this Saga and return it front face up"),
				func(g *game.Game, item *game.StackItem) error {
					greatWorkGrantGraveyardCasts(g, item.Controller, item.SourceCardID)
					return Flicker{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
}

// urabraskOracleID is shared with the back face, which registers under it
// plus "#1" (game.CatalogKey).
const urabraskOracleID = "4b1da9aa-a30c-44b8-a10e-f9dc3fe70b6f"

// greatWorkGrantGraveyardCasts is "until end of turn, you may cast instant
// and sorcery spells from any graveyard. If a spell cast this way would be
// put into a graveyard, exile it instead": one standing permission over each
// seat's graveyard, the way Shaman's Trance opens the others'.
func greatWorkGrantGraveyardCasts(g *game.Game, player, source uuid.UUID) {
	for _, seat := range g.Seats {
		if seat == nil {
			continue
		}
		g.GrantCastPermissionForEffect(game.CastPermission{
			Player:                  player,
			Zone:                    game.ZoneGraveyard,
			Scope:                   game.ScopeStanding,
			ZoneOwner:               seat.ID,
			Filter:                  game.PermissionFilter{InstantOrSorceryOnly: true},
			Duration:                g.UntilEndOfTurnDuration(),
			AltCostKey:              game.AltCostKeyGrantedCost,
			CastOnly:                true,
			ExileInsteadOfGraveyard: true,
			Source:                  source,
			SourceName:              "The Great Work",
			Label:                   "The Great Work — cast an instant or sorcery spell from a graveyard this turn",
		})
	}
}
