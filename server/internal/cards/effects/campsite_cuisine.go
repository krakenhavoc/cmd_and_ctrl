package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Campsite Cuisine — Enchantment {1}{G} (Final Fantasy, #4462):
//
//	"Whenever this enchantment or a legendary creature you control enters,
//	 create a Food token.
//	 Whenever you attack, you may sacrifice X Foods. When you do, up to X
//	 target attacking creatures each get +3/+3 and gain trample and
//	 indestructible until end of turn."
//
// The first trigger is a creature-or-itself entering under your control
// (CR 603.2), written with Do's Food primitive so it declares a footprint.
//
// "Whenever you attack" is once per declaration (CR 508.3d, OncePerBatch).
// "Sacrifice X Foods" is a resolution-time choice among the Foods you
// control, any number from none to all of them (any artifact with the Food
// subtype, a token or not), sacrificed together; X is how many were
// sacrificed, so each Food pays one, and none pays twice (the ruling).
// Sacrificing none is "no", so the reflexive trigger (CR 603.12) is only
// created for one or more. Its target clause is "up to X target attacking
// creatures", X riding the body's Params, and each still-legal pick gets
// +3/+3, trample and indestructible until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45f58002-c831-4b35-8b8c-9a502118e2ec",
		Name:         "Campsite Cuisine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, campsiteCuisineEntered,
				"Campsite Cuisine — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1})),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Campsite Cuisine — you may sacrifice X Foods; when you do, up to X target attacking creatures get +3/+3 and gain trample and indestructible until end of turn",
				campsiteCuisineSacrifice)),
		},
	})
}

// campsiteCuisineEntered is "this enchantment or a legendary creature you
// control enters".
func campsiteCuisineEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && isLegendary(&c)
}

// campsiteCuisineSacrifice asks which Foods to sacrifice.
func campsiteCuisineSacrifice(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
		Question: "Campsite Cuisine — sacrifice X Foods (up to X attacking creatures get +3/+3, trample and indestructible)",
		Keyed:    campsiteCuisineSacrificed,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsArtifact() && c.HasSubtype("Food") {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 0
		},
	}.Apply(NewContext(g, item))
}

// campsiteCuisineSacrificed creates the reflexive trigger for the Foods
// really sacrificed. An on-disk key: never renamed, never reused.
var campsiteCuisineSacrificed = RegisterPermanentsPick("prompt-run/campsite-cuisine-sacrifice-foods", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		n := picked.Count()
		if n == 0 {
			return nil
		}
		t := WhenYouDo("Campsite Cuisine — up to X target attacking creatures each get +3/+3 and gain trample and indestructible until end of turn", campsiteCuisineBoostBody)
		t.Params = game.EffectParams{Amount: n}
		return t.Apply(ctx)
	},
})

// campsiteCuisineBoostBody gives each still-legal target +3/+3, trample and
// indestructible until end of turn. Its clause is "up to X target attacking
// creatures", X being Params.Amount.
var campsiteCuisineBoostBody = game.ReflexiveBody("campsite-cuisine/boost-attackers", simpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Campsite Cuisine — +3/+3"}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample", "indestructible"}, Label: "Campsite Cuisine — trample and indestructible"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}), func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
	return TargetCreature("up to X target attacking creatures", AttackingCreature()).WithCount(0, p.Amount)
})
