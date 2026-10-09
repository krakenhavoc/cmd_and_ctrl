package effects

import (
	"reflect"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// target_purpose_test.go — ADR 0126's amendment of 2026-10-08: a
// purpose says what happens to each target, one entry per target
// clause, and the registration guard holds each entry to its clause.

// targetEntriesOf is a spec's target entries by where they are
// declared: "card", or "mode <i>".
func targetEntriesOf(s Spec) map[string][]game.TargetPurpose {
	out := map[string][]game.TargetPurpose{}
	if ts := s.Purpose.Targets.List(); len(ts) > 0 {
		out["card"] = ts
	}
	if s.Modes != nil {
		for i, o := range s.Modes.Options {
			if ts := o.Purpose.Targets.List(); len(ts) > 0 {
				out["mode "+string(rune('0'+i))] = ts
			}
		}
	}
	return out
}

func TestCuratedTargetPurposeDeclarations(t *testing.T) {
	dmg := func(n int) []game.TargetPurpose { return []game.TargetPurpose{{Slot: 0, Damage: n}} }
	ret := func(r game.TargetReturn) map[string][]game.TargetPurpose {
		return map[string][]game.TargetPurpose{"card": {{Slot: 0, Returns: r}}}
	}
	cases := []struct {
		oracle, name string
		want         map[string][]game.TargetPurpose
	}{
		{"fa3e28b1-131c-4223-81e0-18dfbab22c26", "Prismari Command", map[string][]game.TargetPurpose{
			"mode 0": dmg(2),
			"mode 1": {{Slot: 0, Draws: 2, Discards: 2}},
			"mode 2": {{Slot: 0, Tokens: 1}},
		}},
		{"c6207f6a-a624-4754-88f5-dbe700c841ff", "Sign in Blood", map[string][]game.TargetPurpose{
			"card": {{Slot: 0, Draws: 2, LifeLoss: 2}},
		}},
		{"4457ed35-7c10-48c8-9776-456485fdf070", "Lightning Bolt", map[string][]game.TargetPurpose{"card": dmg(3)}},
		{"a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d", "Shock", map[string][]game.TargetPurpose{"card": dmg(2)}},
		{"f07bd49d-8e71-4d56-be2a-638514011318", "Fiery Temper", map[string][]game.TargetPurpose{"card": dmg(3)}},
		{"f9db72dc-9a5b-48a4-a86e-7464d9a2166a", "Abrade", map[string][]game.TargetPurpose{"mode 0": dmg(3)}},
		{"a07698f6-5ad5-49a3-9da2-f82d407f5cd7", "Izzet Charm", map[string][]game.TargetPurpose{"mode 1": dmg(2)}},
		// #2679: removal that hands its target's controller something.
		{"06692cd9-ac2f-4a32-8fd1-043ba3c0fe71", "Rapid Hybridization", ret(game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3})},
		{"05849bd6-8f38-4031-be2b-e2aa03beb8cc", "Pongify", ret(game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3})},
		{"7735eeba-693b-47e2-bd51-414379cf1016", "Beast Within", ret(game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3})},
		{"fae37e28-e137-4177-b973-fa8b4dd8f409", "Generous Gift", ret(game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3})},
		{"9a107e48-3d50-4941-95b1-10f2b29a4245", "Stroke of Midnight", ret(game.TargetReturn{CreatureTokens: 1, TokenPower: 1, TokenToughness: 1})},
		{"b1544f21-7e98-461b-aed5-e748b0168c52", "Swords to Plowshares", ret(game.TargetReturn{LifeEqualToPower: true})},
		{"d683d985-9888-4d21-8b5f-69e69ce4a03b", "Path to Exile", ret(game.TargetReturn{Lands: 1})},
		{"ac10d218-f9a6-4058-9cda-a15ca1b0b7b5", "Assassin's Trophy", ret(game.TargetReturn{Lands: 1, LandsUntapped: 1})},
	}
	for _, c := range cases {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if got := targetEntriesOf(spec); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s declares target entries %+v, want %+v", c.name, got, c.want)
		}
	}
}

// TestTargetEntriesLeaveTheAmountsAlone: a target entry is the target's,
// so declaring one adds nothing to the controller's amounts. Prismari's
// loot is not "you draw two".
func TestTargetEntriesLeaveTheAmountsAlone(t *testing.T) {
	spec, ok := Lookup("fa3e28b1-131c-4223-81e0-18dfbab22c26")
	if !ok {
		t.Fatal("Prismari Command is not registered")
	}
	for i, o := range spec.Modes.Options {
		p := o.Purpose
		p.Targets = nil
		if !p.IsZero() {
			t.Errorf("Prismari Command mode %d declares controller amounts %+v", i, p)
		}
	}
}

func TestTargetPurposeGuard(t *testing.T) {
	entries := func(ts ...game.TargetPurpose) game.Purpose { return ForTargets(ts...) }
	empty := game.TargetPurposes{}
	cases := []struct {
		name string
		spec Spec
		want string // "" = accepted
	}{
		{"damage to any target", Spec{Targets: TargetAny(), Purpose: entries(DamageToTarget(0, 3))}, ""},
		{"a player gift", Spec{Targets: TargetPlayer("target player"), Purpose: entries(game.TargetPurpose{Draws: 2, LifeLoss: 2})}, ""},
		{"damage to a creature", Spec{Targets: TargetCreature("target creature"), Purpose: entries(DamageToTarget(0, 2))}, ""},
		{"two clauses", Spec{Targets: TargetCreature("target creature").Then(TargetPlayer("target player")),
			Purpose: entries(DamageToTarget(0, 2), game.TargetPurpose{Slot: 1, Draws: 1})}, ""},
		{"on a mode", Spec{Modes: ChooseOne(
			Mode("Draw.", TargetPlayer("target player")),
			ModeWithPurpose(Mode("Burn.", TargetAny()), entries(DamageToTarget(0, 2))),
		)}, ""},
		{"on a triggered row", Spec{Triggered: []game.TriggeredAbility{
			{Key: "t", Targets: TargetAny(), Purpose: entries(DamageToTarget(0, 1))},
		}}, ""},
		{"on an activated row", Spec{Activated: []ActivatedAbility{
			{Label: "a", Targets: TargetPlayer("target player"), Purpose: entries(game.TargetPurpose{Draws: 1})},
		}}, ""},
		{"an empty list", Spec{Targets: TargetAny(), Purpose: game.Purpose{Targets: &empty}}, "empty list"},
		{"no clause", Spec{Purpose: entries(DamageToTarget(0, 3))}, "not one of the statement's 0"},
		{"past the last clause", Spec{Targets: TargetAny(), Purpose: entries(DamageToTarget(1, 3))}, "slot 1"},
		{"a negative slot", Spec{Targets: TargetAny(), Purpose: entries(DamageToTarget(-1, 3))}, "slot -1"},
		{"a slot twice", Spec{Targets: TargetAny(), Purpose: entries(DamageToTarget(0, 3), DamageToTarget(0, 1))}, "twice"},
		{"says nothing", Spec{Targets: TargetAny(), Purpose: entries(game.TargetPurpose{})}, "says nothing"},
		{"negative", Spec{Targets: TargetAny(), Purpose: entries(game.TargetPurpose{Damage: 2, LifeGain: -1})}, "negative amount"},
		{"a draw for a creature", Spec{Targets: TargetCreature("target creature"), Purpose: entries(game.TargetPurpose{Draws: 1})}, "cannot target a player"},
		{"a token for a creature", Spec{Targets: TargetCreature("target creature"), Purpose: entries(game.TargetPurpose{Tokens: 1})}, "cannot target a player"},
		{"damage to a graveyard card", Spec{Targets: TargetCardInGraveyard("target card in a graveyard"), Purpose: entries(DamageToTarget(0, 2))}, "neither a player nor a permanent"},
		{"damage to a spell", Spec{Targets: TargetSpell("target spell"), Purpose: entries(DamageToTarget(0, 2))}, "neither a player nor a permanent"},
		{"on a mode with no clause", Spec{Modes: ChooseOne(
			ModeWithPurpose(Mode("Draw."), entries(game.TargetPurpose{Draws: 1})),
			Mode("Burn.", TargetAny()),
		)}, "spell mode 0"},
		{"on an overload, which clears the clause", Spec{Targets: TargetCreature("target creature"),
			AlternativeCosts: []game.AlternativeCost{CostWithPurpose(Overload("{4}{R}"), entries(DamageToTarget(0, 2)))}},
			"not one of the statement's 0"},
		{"a token back to a creature's controller", Spec{Targets: TargetCreature("target creature"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3}))}, ""},
		{"a land back to a permanent's controller", Spec{Targets: TargetPermanent("target permanent"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{Lands: 1, LandsUntapped: 1}))}, ""},
		{"life back, by power", Spec{Targets: TargetCreature("target creature"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{LifeEqualToPower: true}))}, ""},
		{"a return for a player", Spec{Targets: TargetPlayer("target player"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{Lands: 1}))}, "cannot target a permanent"},
		{"a return for a spell", Spec{Targets: TargetSpell("target spell"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 1, TokenToughness: 1}))}, "cannot target a permanent"},
		{"a token with no size", Spec{Targets: TargetCreature("target creature"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{CreatureTokens: 1}))}, "printed toughness"},
		{"a 3/0 token", Spec{Targets: TargetCreature("target creature"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 3}))}, "printed toughness"},
		{"a size with no token", Spec{Targets: TargetCreature("target creature"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{TokenPower: 3, TokenToughness: 3}))}, "printed toughness"},
		{"more untapped lands than lands", Spec{Targets: TargetCreature("target creature"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{Lands: 1, LandsUntapped: 2}))}, "untapped lands"},
		{"a negative return", Spec{Targets: TargetCreature("target creature"),
			Purpose: entries(RemovalReturning(0, game.TargetReturn{Lands: -1}))}, "negative return"},
		{"on a row built at trigger time", Spec{Triggered: []game.TriggeredAbility{{
			Key: "t", Purpose: entries(DamageToTarget(0, 1)),
			TargetsFrom: func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec { return TargetAny() },
		}}}, "TargetsFrom"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.spec.Name = "Test Card"
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				checkSpecPurposes(c.spec)
			}()
			switch {
			case c.want == "" && msg != "":
				t.Errorf("refused: %s", msg)
			case c.want != "" && !strings.Contains(msg, c.want):
				t.Errorf("got %q, want a refusal containing %q", msg, c.want)
			}
		})
	}
}
