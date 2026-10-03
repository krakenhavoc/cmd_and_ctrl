package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "14c8a590-88ec-4982-ad7b-36a219ff6de7",
		Name:            "Tyvar Kell",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Grants:          []AbilityGrant{TapForManaGrant(tyvarKellGrant, "{B}", "Add {B}", "{T}: Add {B}.")},
		Static: []game.StaticAbility{
			GrantAbilities(tyvarKellElvesYouControl, tyvarKellGrant),
		},
		Emblem: &EmblemSpec{
			Label: "Tyvar Kell emblem",
			Text:  "Whenever you cast an Elf spell, it gains haste until end of turn and you draw two cards.",
			Triggered: []game.TriggeredAbility{
				WheneverYouCast(HasSubtype("Elf"), "Tyvar Kell emblem — it gains haste, and you draw two cards",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if err := (ThatSpellGains{Keywords: []string{"haste"}, UntilEndOfTurn: true}).Apply(ctx); err != nil {
							return err
						}
						return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
					}),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Put a +1/+1 counter on up to one target Elf. Untap it. It gains deathtouch until end of turn.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target Elf", HasSubtype("Elf")).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
						if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"deathtouch"}}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label: "0: Create a 1/1 green Elf Warrior creature token.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 green Elf Warrior"), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−6: You get an emblem with \"Whenever you cast an Elf spell, it gains haste until end of turn and you draw two cards.\"",
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
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
	return target.Controller == source.Controller && target.HasSubtype("Elf")
}
