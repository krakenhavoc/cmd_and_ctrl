package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_types.go is ADR 0109 §1's card-side vocabulary (#1881): a land
// that becomes a basic land type for a while.
//
//	"Target land becomes an Island until end of turn."      Tidal Warrior
//	"Target land becomes the basic land type of your choice
//	 until end of turn."                                    Reef Shaman
//	"Choose a basic land type. Each land you control
//	 becomes that type until end of turn."                  Terraformer
//	"Until end of turn, target land you control becomes the
//	 basic land type of your choice in addition to its
//	 other types."                                          Navigator's Compass
//
// "Becomes a Forest" is CR 305.7's type SET: the land's old land types
// and the abilities its rules text gives it go, its other subtypes stay
// (CR 205.1a), and it taps for the new colour. That is the engine's
// game.ModSetBasicLandTypes, a data record, so a table holding one is a
// restore point. "In addition to its other types" takes nothing away
// (CR 205.1b) and is game.ModAddSubtypes. The static form, "enchanted
// land is an Island", is SetsBasicLandType (attachments.go).
//
// The set a group effect changes is fixed as it begins (CR 611.2c), so
// a land that arrives later in the turn is not a Swamp.

// LandBecomes is "<land> becomes <types>" from a resolving spell or
// ability: Target or Match picks the lands, exactly as ScopedEffectFor
// does, and Types are the land types they become.
type LandBecomes struct {
	// Target pins the effect to one land. Ignored when Match is set.
	Target uuid.UUID
	// Match selects the affected lands, once, now (CR 611.2c).
	Match CardPredicate
	// Types are the land types. Without InAddition each must be a basic
	// land type (CR 305.7); registration refuses anything else.
	Types []string
	// InAddition is "in addition to its other types" (CR 205.1b).
	InAddition bool
	// Duration is how long it lasts (CR 611.2a). Build it with one of
	// the Duration* builders in durations.go.
	Duration game.Duration
	// Label is attribution for logs and tests.
	Label string
}

// Apply registers the record. Nothing matched registers nothing.
func (l LandBecomes) Apply(ctx *Context) error {
	mod := game.SetBasicLandTypesMod(l.Types...)
	if l.InAddition {
		mod = game.AddSubtypesMod(l.Types...)
	}
	return ScopedEffectFor{
		Target:   l.Target,
		Match:    l.Match,
		Mods:     []game.Mod{mod},
		Duration: l.Duration,
		Label:    l.Label,
	}.Apply(ctx)
}

// ChooseBasicLandTypeThen asks the resolving effect's controller to
// choose one of `options` (CR 305.6's five when nil, Tundra Kavu's
// "a Plains or an Island" when two) and hands the answer to `then`. The
// choice is made as the effect resolves (CR 608.2), through the
// ordinary option pick, so the bot and the client answer it unchanged.
// `then` is not run when no question could be asked.
//
// Vision Charm's "a land type" is the same call over game.LandTypes.
func ChooseBasicLandTypeThen(ctx *Context, question string, options []string, then func(ctx *Context, landType string) error) error {
	if len(options) == 0 {
		options = game.BasicLandTypes
	}
	words := append([]string(nil), options...)
	choices := make([]game.ChoiceOption, len(words))
	for i, w := range words {
		choices[i] = game.ChoiceOption{Label: w}
	}
	return PickOption{
		Question: question,
		Options:  choices,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(words) {
				return nil
			}
			return then(ctx, words[index])
		},
	}.Apply(ctx)
}

// landTypeArticle is "an Island", "a Forest": the type with the article
// a card prints before it.
func landTypeArticle(t string) string {
	if t != "" && strings.ContainsRune("AEIOU", rune(t[0])) {
		return "an " + t
	}
	return "a " + t
}

// landBecomesOrChoose is the shared body: with one type, the land
// becomes it; with several (or none, meaning all five basic land types)
// the controller chooses among them first. The land is the first legal
// battlefield target, and a target gone by resolution changes nothing
// and asks nothing (CR 608.2b).
//
// `after` is the rest of the spell's text, run once the land has its
// type (CR 608.2c: in the order written), so Shimmering Mirage's "Draw
// a card" comes after the choice and not before it. Nil for none.
func landBecomesOrChoose(ctx *Context, name string, dur func(ctx *Context, land uuid.UUID) (game.Duration, bool), inAddition bool, types []string, after func(ctx *Context) error) error {
	then := func(ctx *Context) error {
		if after == nil {
			return nil
		}
		return after(ctx)
	}
	land := FirstLegalBattlefieldTarget(ctx)
	if land == uuid.Nil {
		return then(ctx)
	}
	become := func(ctx *Context, t string) error {
		if d, ok := dur(ctx, land); ok {
			label := name + " — that land is " + landTypeArticle(t)
			if inAddition {
				label += " in addition to its other types"
			}
			if err := (LandBecomes{Target: land, Types: []string{t}, InAddition: inAddition, Duration: d, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		return then(ctx)
	}
	if len(types) == 1 {
		return become(ctx, types[0])
	}
	question := name + " — choose a basic land type"
	if len(types) > 1 {
		question = name + " — choose " + joinOr(types)
	}
	return ChooseBasicLandTypeThen(ctx, question, types, become)
}

// untilEndOfTurnFor is the duration "until end of turn" in the shape
// landBecomesOrChoose takes.
func untilEndOfTurnFor(ctx *Context, _ uuid.UUID) (game.Duration, bool) {
	return DurationUntilEndOfTurn(ctx), true
}

// TargetLandBecomesUntilEOT is the Effect of "Target land becomes
// <a type> until end of turn": a fixed type with one argument (Tidal
// Warrior's Island), a choice among several (Tundra Kavu's "a Plains or
// an Island"), and "the basic land type of your choice" with none
// (Reef Shaman). The ability's Targets clause says which lands may be
// chosen.
func TargetLandBecomesUntilEOT(name string, types ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return landBecomesOrChoose(NewContext(g, item), name, untilEndOfTurnFor, false, types, nil)
	}
}

// TargetLandBecomesWhileSourceRemains is "target land becomes a <type>
// until this creature leaves the battlefield" (Gaea's Liege) and "for
// as long as this creature remains on the battlefield" (Tide Shaper):
// the same CR 611.2b duration. A source that has already left as the
// ability resolves makes the effect never begin, so nothing happens.
func TargetLandBecomesWhileSourceRemains(name, landType string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return landBecomesOrChoose(ctx, name, func(ctx *Context, land uuid.UUID) (game.Duration, bool) {
			d, ok := DurationWhileSourceRemains(ctx, item.SourceCardID)
			// The pin only collects the record once the land is gone
			// (CR 400.7); the source's condition is what ends it.
			return ctx.Game.PinnedTo(d, land), ok
		}, false, []string{landType}, nil)
	}
}

// TargetLandBecomesIndefinitely is "target land becomes a <type>" with
// no duration printed (CR 611.2a: it lasts until the game ends), pinned
// to the land so it ends when the land stops being that object (CR
// 400.7): Thelonite Monk, Cyclopean Giant.
func TargetLandBecomesIndefinitely(name, landType string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return landBecomesOrChoose(NewContext(g, item), name, func(ctx *Context, land uuid.UUID) (game.Duration, bool) {
			return ctx.Game.PinnedTo(game.IndefiniteDuration(), land), true
		}, false, []string{landType}, nil)
	}
}

// TargetLandBecomesUntilItsControllersNextTurn is "target land becomes
// a <type> until its controller's next untap step" (Orcish Farmer). The
// untap step is the first step of a turn (CR 501.1), and an "until
// your next turn" duration ends as that turn begins, before anything
// untaps, so the two are the same moment. "Its controller" is the land's
// controller as the ability resolves.
func TargetLandBecomesUntilItsControllersNextTurn(name, landType string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return landBecomesOrChoose(NewContext(g, item), name, func(ctx *Context, land uuid.UUID) (game.Duration, bool) {
			c, ok := ctx.Game.LookupCardForEffect(land)
			if !ok {
				return game.Duration{}, false
			}
			return ctx.Game.PinnedTo(DurationUntilYourNextTurn(ctx, c.Controller), land), true
		}, false, []string{landType}, nil)
	}
}

// eachLandYouControlBecomesChosenTypeUntilEOT is "Choose a basic land
// type. Each land you control becomes that type until end of turn"
// (Terraformer, Elsewhere Flask). The lands are the ones you control as
// it resolves, after the choice (CR 611.2c).
func eachLandYouControlBecomesChosenTypeUntilEOT(ctx *Context, name string) error {
	return ChooseBasicLandTypeThen(ctx, name+" — choose a basic land type", nil, func(ctx *Context, t string) error {
		return LandBecomes{
			Match:    And(Land(), YouControl()),
			Types:    []string{t},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    name + " — each land you control is " + landTypeArticle(t),
		}.Apply(ctx)
	})
}

// TargetLandBecomesChosenTypeThen is the spell form of "Target land
// becomes the basic land type of your choice until end of turn", with
// the rest of the spell's text in `after` (Shimmering Mirage's "Draw a
// card", Jinx's delayed draw).
func TargetLandBecomesChosenTypeThen(ctx *Context, name string, after func(ctx *Context) error) error {
	return landBecomesOrChoose(ctx, name, untilEndOfTurnFor, false, nil, after)
}

// TargetLandYouControlGainsChosenTypeUntilEOT is "Until end of turn,
// target land you control becomes the basic land type of your choice in
// addition to its other types" (Navigator's Compass): the land keeps its
// land types and rules text and gains the chosen one, with its mana
// ability (CR 205.1b, 305.7's last sentence).
func TargetLandYouControlGainsChosenTypeUntilEOT(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return landBecomesOrChoose(NewContext(g, item), name, untilEndOfTurnFor, true, nil, nil)
	}
}

// ChooseBasicLandTypeAsEnters is "As this Aura enters, choose a basic
// land type" (CR 614.12): Convincing Mirage, Phantasmal Terrain. The
// answer lands on the permanent's Card.ChosenOption, the anchor-word
// store (choose_option.go), so EnchantedLandIsTheChosenType's five gated
// statics read it.
func ChooseBasicLandTypeAsEnters(name string) func(*game.Card, *Context) error {
	return ChooseOptionAsEnters(name, game.BasicLandTypes...)
}

// EnchantedLandIsTheChosenType is "Enchanted land is the chosen type":
// one SetsBasicLandType per basic land type, each gated on that type
// being the Aura's chosen option (ChosenIs), so exactly one applies once
// the choice is made and none before.
func EnchantedLandIsTheChosenType() []game.StaticAbility {
	out := make([]game.StaticAbility, 0, len(game.BasicLandTypes))
	for _, t := range game.BasicLandTypes {
		out = append(out, StaticWhenChosen(t, SetsBasicLandType(AttachedToSource, nil, []string{t})))
	}
	return out
}

// ---------------------------------------------------------------
// "Loses all land types" (ADR 0109 §2, #1604)
// ---------------------------------------------------------------
//
//	"Enchanted land loses all land types and abilities and has
//	 '{T}: Add {C}' and '{T}, Pay 1 life: Add one mana of any
//	 color.'"                                       Lithoform Blight
//	"Lands your opponents control with the chosen name lose all land
//	 types and abilities, and they gain '{T}: Add one mana of any
//	 color.'"                                            Alpine Moon
//
// Every printed use is one sentence in three layers, and each layer is
// a separate StaticAbility, declared in the order the card prints them
// (the shape Imprisoned in the Moon uses):
//
//   - layer 4: LosesAllLandTypes. Every CR 205.3i land type goes and
//     every other subtype stays (CR 205.1a), so a Dryad Arbor keeps
//     its Dryad. This is NOT CR 305.7, which is about setting a basic
//     land type: it takes no ability by itself. What it does take is
//     the intrinsic mana ability of each basic land type the land had
//     (CR 305.6), because that is derived from the effective subtypes.
//   - layer 6: LosesAllAbilitiesAndHas. "Loses all abilities" is the
//     CR 613.1f removal, and the abilities it "has" are the same
//     effect's grant, written after the removal empties the list (ADR
//     0093 Decision 3), so the granted mana abilities are all the land
//     has. Unlike CR 305.7's layer-4 removal, a layer-6 removal also
//     takes an ability another effect granted earlier (CR 613.7).
//
// The resolved-effect form is the record game.LoseLandTypesMod (Ultima,
// Origin of Oblivion), with game.LoseAllAbilitiesMod and a grant in the
// same record.

// LosesAllLandTypes is the layer-4 static "<affected> loses all land
// types" over every permanent `applies` matches.
func LosesAllLandTypes(applies func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer4Type,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.LoseLandTypes()
		},
	}
}

// LosesAllAbilitiesAndHas is the layer-6 half, "loses all abilities and
// has '<ability>'": LoseAllAbilities over `applies`, with the named
// ability bundles (Spec.Grants) as the same effect's grant. It keeps
// LoseAllAbilities' ContinuesAfterRemoval, because every printed use is
// one sentence with a layer-4 change (CR 613.6).
func LosesAllAbilitiesAndHas(applies func(target *game.Card, g *game.Game, source *game.Card) bool, keys ...string) game.StaticAbility {
	s := LoseAllAbilities()
	s.AppliesTo = applies
	s.GrantAbilities = append([]string(nil), keys...)
	return s
}
