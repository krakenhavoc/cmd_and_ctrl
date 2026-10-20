package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leori, Sparktouched Hunter — Legendary Creature — Elemental Cat
// {U}{R}{W}, 3/3:
//
//	"Flying, vigilance
//	 Whenever Leori deals combat damage to a player, choose a
//	 planeswalker type. Until end of turn, whenever you activate an
//	 ability of a planeswalker of that type, copy that ability. You may
//	 choose new targets for the copies."
//
// The choice is made as the trigger resolves (CR 608.2), among CR
// 205.3j's planeswalker types, through the ordinary option pick; the
// types of the planeswalkers you control are offered first. The answer
// sets up a repeating CR 603.7 delayed trigger for the rest of the turn
// (Great Train Heist's "whenever … this turn", CR 603.7b): each ability
// you activate of a planeswalker of the chosen type, loyalty or not, is
// copied while it is on the stack (CR 707.10), with the new-targets
// offer. The planeswalker is read as the ability is activated, so a
// walker of that type that arrives later this turn counts. A mana
// ability never uses the stack and is never copied (CR 605.3b), and a
// copy is not activated, so it is not copied again (CR 707.10).
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	const name = "Leori, Sparktouched Hunter"
	Register(Spec{
		OracleID:        "ff94e2d5-8ba1-4f6c-916d-58235a60743e",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(name+" — choose a planeswalker type; copy its abilities this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					types := leoriPlaneswalkerTypeOptions(g, item.Controller)
					options := make([]game.ChoiceOption, len(types))
					for i, t := range types {
						options[i] = game.ChoiceOption{Label: t}
					}
					return PickOption{
						Question: name + " — choose a planeswalker type",
						Options:  options,
						ThenKey:  leoriChoseAType,
					}.Apply(ctx)
				}),
		},
	})
}

var (
	// The answer: until end of turn, copy each ability you activate of a
	// planeswalker of the chosen type. On-disk identity.
	leoriChoseAType = OptionPickThen("option-pick/leori-sparktouched-hunter-planeswalker-type",
		func(ctx *Context, r game.OptionPicked) error {
			if r.Option == nil || r.Option.Label == "" {
				return nil
			}
			pwType := r.Option.Label
			return DelayedOnEvent{
				Label:      "Leori, Sparktouched Hunter — copy that " + pwType + " ability",
				On:         []game.EventKind{game.EventActivateAbility},
				Condition:  leoriActivatedAbilityOfTypeCondition,
				CondParams: game.EffectParams{Name: pwType},
				Body:       copyTheAbilityBody,
				Repeats:    true,
			}.Apply(ctx)
		})

	// "Whenever you activate an ability of a planeswalker of that type":
	// the activation's source, as it is now, is a planeswalker with the
	// subtype in CondParams.Name. On-disk identity.
	leoriActivatedAbilityOfTypeCondition = game.DelayedCondition("activate/you-activate-an-ability-of-a-planeswalker-type",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
			if ev.Kind != game.EventActivateAbility || ev.Actor != dt.Controller || p.Name == "" {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.IsPlaneswalker() && c.HasSubtype(p.Name)
		})
)

// leoriPlaneswalkerTypeOptions is CR 205.3j's planeswalker types, those
// of the planeswalkers `you` control first, each once.
func leoriPlaneswalkerTypeOptions(g *game.Game, you uuid.UUID) []string {
	all := game.PlaneswalkerTypes()
	var first, rest []string
	for _, t := range all {
		yours := false
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == you && c.IsPlaneswalker() && c.HasSubtype(t) {
				yours = true
				break
			}
		}
		if yours {
			first = append(first, t)
		} else {
			rest = append(rest, t)
		}
	}
	return append(first, rest...)
}
