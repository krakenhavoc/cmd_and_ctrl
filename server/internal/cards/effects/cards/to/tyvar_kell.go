package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyvar Kell — Legendary Planeswalker — Tyvar {2}{G}{G}, loyalty 3:
//
//	"Elves you control have "{T}: Add {B}."
//	 +1: Put a +1/+1 counter on up to one target Elf. Untap it. It
//	     gains deathtouch until end of turn.
//	 0: Create a 1/1 green Elf Warrior creature token.
//	 −6: You get an emblem with "Whenever you cast an Elf spell, it
//	     gains haste until end of turn and you draw two cards.""
//
// The static is an ADR 0093 ability grant: every Elf the controller
// controls has the "{T}: Add {B}." bundle, a creature's subject to
// summoning sickness (CR 302.6).
//
// The emblem's haste is given to the Elf SPELL on the stack, until end
// of turn (ADR 0109 §11 decision 3, #1552): ThatSpellGains writes the
// grant pinned to the spell, the creature it becomes keeps it
// (CR 400.7a), and it ends in that turn's cleanup step. An Elf spell
// countered before the trigger resolves gains nothing, and the two
// cards are drawn either way.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "14c8a590-88ec-4982-ad7b-36a219ff6de7",
		Name:            "Tyvar Kell",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Grants:          []effects.AbilityGrant{effects.TapForManaGrant(tyvarKellGrant, "{B}", "Add {B}", "{T}: Add {B}.")},
		Static: []game.StaticAbility{
			effects.GrantAbilities(tyvarKellElvesYouControl, tyvarKellGrant),
		},
		Emblem: &effects.EmblemSpec{
			Label: "Tyvar Kell emblem",
			Text:  "Whenever you cast an Elf spell, it gains haste until end of turn and you draw two cards.",
			Triggered: []game.TriggeredAbility{
				effects.WheneverYouCast(effects.HasSubtype("Elf"), "Tyvar Kell emblem — it gains haste, and you draw two cards",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if err := (effects.ThatSpellGains{Keywords: []string{"haste"}, UntilEndOfTurn: true}).Apply(ctx); err != nil {
							return err
						}
						return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
					}),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Put a +1/+1 counter on up to one target Elf. Untap it. It gains deathtouch until end of turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target Elf", effects.HasSubtype("Elf")).WithCount(0, 1),
				// The untap and deathtouch wait for a prompt the counter asks (CR 608.2c, #4155).
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
								return err
							}
							if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"deathtouch"}}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
				),
			},
			{
				Label: "0: Create a 1/1 green Elf Warrior creature token.",
				Cost:  effects.LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label: "−6: You get an emblem with \"Whenever you cast an Elf spell, it gains haste until end of turn and you draw two cards.\"",
				Cost:  effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

const tyvarKellGrant = "tyvar-kell/tap-for-black"

// tyvarKellElvesYouControl is "Elves you control": every Elf permanent,
// a Kindred Elf enchantment included — not only Elf creatures, which is
// all TribeFilter reads.
func tyvarKellElvesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	// "Elves" is a word of the source's text (CR 612.2, #3984).
	return target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Elf"))
}
