package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Megatron, Tyrant // Megatron, Destructive Force (#3195, ADR 0152) — a
// More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {3}{R}{W}{B}, 7/5:
//
//	"More Than Meets the Eye {1}{R}{W}{B}
//	 Your opponents can't cast spells during combat.
//	 At the beginning of each of your postcombat main phases, you may
//	 convert Megatron. If you do, add {C} for each 1 life your opponents
//	 have lost this turn."
//
// Back face, Legendary Artifact — Vehicle, 4/5:
//
//	"Living metal
//	 Whenever Megatron attacks, you may sacrifice another artifact. When
//	 you do, Megatron deals damage equal to the sacrificed artifact's
//	 mana value to target creature. If excess damage would be dealt to
//	 that creature this way, instead that damage is dealt to that
//	 creature's controller and you convert Megatron."
//
// "During combat" is every step from the beginning of combat to the end
// of combat step on any player's turn, first-strike damage included. The
// {C} is counted when the "you may" is answered "yes" and Megatron
// really converted (CR 701.28e), over every opponent's life lost this
// turn together. The back face is Ziatora's shape: the sacrifice is made
// as the trigger resolves, and "when you do" is a reflexive trigger that
// chooses its target afterwards. The artifact's mana value is read
// before it is sacrificed and carried as the trigger's amount. Excess is
// split at the lethal line (CR 120.4a) the way Ravenous Tyrannosaurus
// does: the creature takes what it still needed to die, the rest goes to
// its controller, and Megatron converts only when there was any.
//
// No simplification.
const megatronOracleID = "33e88072-3840-4090-8d0b-7fb38279e150"

var megatronStrikeBody = game.ReflexiveBody("megatron-destructive-force/strike", megatronStrike,
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))

func init() {
	Register(Spec{
		OracleID:         megatronOracleID,
		Name:             "Megatron, Tyrant",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{1}{R}{W}{B}")},
		CastRestrictions: []game.CastRestriction{
			{
				Label: "Megatron, Tyrant — your opponents can't cast spells during combat.",
				Forbids: func(q game.CastQuery) bool {
					return q.Source.Controller != q.Controller && inCombatStep(q.Game.Turn.Step)
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			AtYourPostcombatMain("Megatron, Tyrant — you may convert Megatron; if you do, add {C} for each life your opponents lost this turn",
				megatronPostcombat),
		},
	})
	Register(Spec{
		OracleID:     megatronOracleID + "#1",
		Name:         "Megatron, Destructive Force",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{LivingMetal()},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Megatron, Destructive Force — you may sacrifice another artifact; when you do, damage to target creature",
				megatronMaySacrifice),
		},
	})
}

// inCombatStep is "during combat": the combat phase's steps. It reads
// the step it is handed, so it needs no lock (the cast path holds g.mu).
func inCombatStep(s game.Step) bool {
	switch s {
	case game.StepBeginCombat, game.StepDeclareAttackers, game.StepDeclareBlockers,
		game.StepFirstStrikeDamage, game.StepCombatDamage, game.StepEndCombat:
		return true
	}
	return false
}

func megatronPostcombat(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Megatron, Tyrant — convert Megatron? (If you do, add {C} for each life your opponents have lost this turn.)",
		Then:     megatronConvertAnswer,
	}.Apply(NewContext(g, item))
}

func megatronMaySacrifice(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	return ChoosePermanents{
		Question: "Megatron, Destructive Force — sacrifice another artifact to deal damage equal to its mana value to target creature?",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, id := range permanentsControlledByMatching(g, of, Artifact()) {
				if id != self {
					out = append(out, id)
				}
			}
			return out, 0, 1
		},
		Keyed: megatronSacrificed,
	}.Apply(NewContext(g, item))
}

// megatronSacrificed is the chosen artifact sacrificed and the reflexive
// damage raised once it has gone, keyed (#4290).
var megatronSacrificed = RegisterPermanentsPick("prompt-run/megatron-tyrant-sacrifice", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		cards := picked.Cards()
		if len(cards) == 0 {
			return nil
		}
		mv := 0
		if c, ok := ctx.Game.LookupCardForEffect(cards[0]); ok {
			mv, _ = ctx.Game.ManaValueForEffect(c)
		}
		item := ctx.Item
		return ctx.Game.SacrificeAllThenForEffect(uuid.Nil, cards, func(g *game.Game, sacrificed []uuid.UUID) error {
			if len(sacrificed) == 0 {
				return nil
			}
			t := WhenYouDo("Megatron, Destructive Force — damage equal to the sacrificed artifact's mana value to target creature", megatronStrikeBody)
			t.Params = game.EffectParams{Amount: mv}
			return t.Apply(NewContext(g, item))
		})
	},
})

// megatronStrike is the reflexive trigger's body: Megatron deals the
// carried mana value to the target creature, split at its lethal line;
// the excess goes to the creature's controller instead, and Megatron
// converts.
func megatronStrike(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := NewContext(g, item)
	amount := p.Amount
	if amount <= 0 {
		return nil
	}
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
			break
		}
	}
	victim, ok := g.LookupCardForEffect(target)
	if target == uuid.Nil || !ok {
		return nil
	}
	lethal := max(victim.CurrentToughness()-victim.DamageMarked, 0)
	toCreature, excess := amount, 0
	if amount > lethal {
		toCreature, excess = lethal, amount-lethal
	}
	if err := (DealDamage{Source: item.SourceCardID, Target: target, Amount: toCreature}).Apply(ctx); err != nil {
		return err
	}
	if excess <= 0 {
		return nil
	}
	if err := (DealDamage{Source: item.SourceCardID, Target: victim.Controller, Amount: excess}).Apply(ctx); err != nil {
		return err
	}
	return ConvertThis{}.Apply(ctx)
}

// megatronConvertAnswer is the "yes": convert Megatron, then add {C} for
// each life the opponents have lost this turn. Keyed (#4291).
var megatronConvertAnswer = MayYes("confirm/megatron-tyrant-convert", func(ctx *Context, _ game.ConfirmAnswer) error {
	converted, err := convertThisReporting(ctx)
	if err != nil || !converted {
		return err
	}
	lost := 0
	for _, opp := range ctx.Opponents() {
		lost += b18LifeLostThisTurn(ctx.Game, opp)
	}
	if lost <= 0 {
		return nil
	}
	return AddMana{Amount: ManaN(lost, "C")}.Apply(ctx)
})
