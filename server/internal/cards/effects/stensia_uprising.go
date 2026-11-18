package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stensia Uprising — Enchantment {2}{R}{R} (Innistrad: Crimson Vow):
//
//	"At the beginning of your end step, create a 1/1 red Human creature
//	 token. Then if you control exactly thirteen permanents, you may
//	 sacrifice this enchantment. When you do, it deals 7 damage to any
//	 target."
//
// One trigger. The token is made first, and the permanents are counted after
// it exists ("then if", CR 608.2c: not an intervening "if", so it is read as
// the ability resolves, never as it triggers). Exactly thirteen is counted
// across everything the controller controls, tokens and lands included. The
// sacrifice is a "may", asked only when the count is right; the CR 603.12
// reflexive trigger exists only on the branch where the enchantment really
// was sacrificed, and picks its "any target" as it goes on the stack. The
// damage is dealt by the enchantment as it last existed (CR 608.2h), so it
// still lands from the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a881bf6e-6247-4adb-9f9c-e4b928692573",
		Name:         "Stensia Uprising",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Stensia Uprising — create a 1/1 red Human creature token", Do(
				CreateToken{Template: TokenCard("1/1 red Human"), N: 1},
				Instruction(stensiaUprisingMaySacrifice),
			)),
		},
	})
}

// stensiaUprisingMaySacrifice is "then if you control exactly thirteen
// permanents, you may sacrifice this enchantment".
func stensiaUprisingMaySacrifice(ctx *Context) error {
	count := countControlled(ctx.Game, ctx.Controller(), func(game.Card) bool { return true })
	if count != 13 {
		return nil
	}
	return MayChoice{
		Question: "Stensia Uprising — sacrifice it to deal 7 damage to any target?",
		YesLabel: "Sacrifice Stensia Uprising",
		NoLabel:  "Keep it",
		Then:     stensiaUprisingAnswer,
	}.Apply(ctx)
}

// stensiaUprisingAnswer is the "yes": sacrifice the enchantment, then raise
// the reflexive trigger. Keyed (#4291).
var stensiaUprisingAnswer = MayYes("confirm/stensia-uprising-sacrifice", func(ctx *Context, _ game.ConfirmAnswer) error {
	source := ctx.Source()
	// An enchantment that left and came back is not "this enchantment":
	// nothing is sacrificed, so there is no "when you do" (#1432).
	if source == uuid.Nil || ctx.isNewSourceObjectAsThis(source) {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(source); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	if err := ctx.Game.SacrificePermanentForEffect(source); err != nil {
		return err
	}
	return ReflexiveTrigger{
		Label: "Stensia Uprising — 7 damage to any target",
		Body:  stensiaUprisingDamageBody,
	}.Apply(ctx)
})

// stensiaUprisingDamageBody is the reflexive trigger's registration: 7
// damage from the sacrificed enchantment to the target picked as it
// triggered.
var stensiaUprisingDamageBody = game.ReflexiveBody("stensia-uprising/damage", simpleBody(stensiaUprisingDamage),
	func(_ uuid.UUID, _ game.EffectParams) *game.TargetSpec { return TargetAny() })

func stensiaUprisingDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := firstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 7}.Apply(ctx)
}
