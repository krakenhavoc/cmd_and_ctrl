package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b788_helpers.go — Visions, S82 older sets batch 788 (#5631):
// bodies the batch's cards share. Append-only.

// S82OldB788EachLandIsAlso is "Each land is a <subtype> in addition to its
// other land types" (Blanket of Night): one Layer-4 static over every land
// on the battlefield, whoever controls it. The subtype is a word of the
// source's text (CR 612.2), appended so a printed land type keeps its slot.
func S82OldB788EachLandIsAlso(subtype, label string) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer4Type,
		Label: label,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsLand()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			word := source.TextSubtype(subtype)
			for _, st := range c.Subtypes {
				if st == word {
					return
				}
			}
			c.Subtypes = append(c.Subtypes, word)
		},
	}
}

// S82OldB788PlayersActivateOnlyOnTheirOwnTurns is "Players can ... activate
// abilities only during their own turns" (City of Solitude): the activator
// must be the active player. It binds every player, the source's controller
// included, reaches mana abilities (the card makes no exemption) and
// abilities activated from any zone.
func S82OldB788PlayersActivateOnlyOnTheirOwnTurns(label string) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			return !IsActivePlayer(q.Game, q.Controller)
		},
	}
}

// --- Dream Tides --------------------------------------------------------

// S82OldB788DreamTidesUpkeep is Dream Tides' upkeep trigger: "that player may
// choose any number of tapped nongreen creatures they control and pay {2} for
// each creature chosen this way. If the player does, untap those creatures."
// The player whose upkeep it is chooses, then is offered the total payment.
func S82OldB788DreamTidesUpkeep(g *game.Game, item *game.StackItem) error {
	player := TriggeringActor(item)
	if player == uuid.Nil {
		return nil
	}
	var tapped []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.Tapped && !c.HasColorWord("G") {
			tapped = append(tapped, c.InstanceID)
		}
	}
	if len(tapped) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Dream Tides — choose any number of tapped nongreen creatures you control ({2} each)",
		Cards:    tapped,
		Min:      0,
		Max:      len(tapped),
		ThenKey:  s82OldB788DreamTidesPick,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB788DreamTidesPick offers the {2}-per-creature payment for the
// creatures chosen. An on-disk key: never renamed, never reused.
var s82OldB788DreamTidesPick = AskedPick("choose-cards/s82-old-b788-dream-tides-choose",
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		objs := make([]game.ObjectRef, 0, len(r.Picked))
		for _, id := range r.Picked {
			if c, ok := ctx.Game.LookupCardForEffect(id); ok {
				objs = append(objs, game.ObjectRef{ID: id, Epoch: c.ObjectEpoch})
			}
		}
		if len(objs) == 0 {
			return nil
		}
		cost := "{" + strconv.Itoa(2*len(objs)) + "}"
		return MayPay{
			Chooser:      r.Chooser,
			Cost:         cost,
			Question:     "Dream Tides — pay " + cost + " to untap the chosen creatures?",
			InThisStep:   true,
			Then:         s82OldB788DreamTidesPaid,
			CarryObjects: objs,
		}.Apply(ctx)
	})

// s82OldB788DreamTidesPaid untaps the chosen creatures that are still the
// objects that were chosen. An on-disk key: never renamed, never reused.
var s82OldB788DreamTidesPaid = PayUnlessBranches("pay-unless/s82-old-b788-dream-tides-untap",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		for _, o := range r.Objects {
			c, ok := ctx.Game.LookupCardForEffect(o.ID)
			if !ok || c.ObjectEpoch != o.Epoch || !OnBattlefield(ctx.Game, o.ID) {
				continue
			}
			if err := (UntapTarget{Target: o.ID}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}, nil)

// --- Equipoise ----------------------------------------------------------

// equipoiseKinds is the order of the three passes the card prints: lands,
// then artifacts, then creatures.
var equipoiseKinds = []CardPredicate{Land(), Artifact(), Creature()}

// S82OldB788EquipoiseUpkeep is Equipoise's upkeep trigger: "for each land
// target player controls in excess of the number you control, choose a land
// that player controls, then the chosen permanents phase out. Repeat this
// process for artifacts and creatures." Each pass counts when it is reached,
// the ability's controller chooses (CR 608.2d), and a pass whose choice is
// forced phases the permanents out without a prompt.
func S82OldB788EquipoiseUpkeep(g *game.Game, item *game.StackItem) error {
	return equipoisePass(NewContext(g, item), 0)
}

func equipoisePass(ctx *Context, from int) error {
	target, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	me := ctx.Controller()
	for pass := from; pass < len(equipoiseKinds); pass++ {
		theirs := PermanentsControlledByMatching(ctx.Game, target, equipoiseKinds[pass])
		mine := PermanentsControlledByMatching(ctx.Game, me, equipoiseKinds[pass])
		excess := len(theirs) - len(mine)
		if excess <= 0 {
			continue
		}
		if excess >= len(theirs) {
			if err := (PhaseOut{Targets: theirs}).Apply(ctx); err != nil {
				return err
			}
			continue
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   me,
			Source:    ctx.Source(),
			Question:  "Equipoise — choose " + strconv.Itoa(excess) + " " + equipoiseWord(pass) + " to phase out",
			Cards:     theirs,
			Min:       excess,
			Max:       excess,
			ThenKey:   s82OldB788EquipoisePick,
			LiveZone:  game.ZoneBattlefield,
			CarryNums: []int{pass},
			Asked:     game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return nil
}

func equipoiseWord(pass int) string {
	switch pass {
	case 0:
		return "land(s)"
	case 1:
		return "artifact(s)"
	}
	return "creature(s)"
}

// s82OldB788EquipoisePick phases out the permanents chosen in one pass and
// moves on to the next. An on-disk key: never renamed, never reused.
var s82OldB788EquipoisePick = AskedPick("choose-cards/s82-old-b788-equipoise-phase-out",
	func(ctx *Context, r game.CardsChosen) error {
		if err := (PhaseOut{Targets: r.Picked}).Apply(ctx); err != nil {
			return err
		}
		next := 1
		if len(r.Nums) > 0 {
			next = r.Nums[0] + 1
		}
		return equipoiseNextPass(ctx, next)
	})

// equipoiseNextPass is equipoisePass, reached through a variable so the
// registered pick above does not make an initialization cycle with it.
var equipoiseNextPass func(ctx *Context, from int) error

var _ = func() struct{} { equipoiseNextPass = equipoisePass; return struct{}{} }()

// --- Guiding Spirit -----------------------------------------------------

// S82OldB788GuidingSpiritEffect is "If the top card of target player's
// graveyard is a creature card, put that card on top of that player's
// library." The condition is read as the ability resolves.
func S82OldB788GuidingSpiritEffect(g *game.Game, item *game.StackItem) error {
	player, ok := FirstLegalPlayerTarget(NewContext(g, item))
	if !ok {
		return nil
	}
	top := g.TopGraveyardCardsForEffect(player, 1, nil)
	if len(top) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(top[0])
	if !ok || !c.IsCreature() {
		return nil
	}
	return g.TuckToLibraryForEffect(top[0], false)
}

// --- Quirion Druid ------------------------------------------------------

// S82OldB788AnimateTargetLand is "Target land becomes a 2/2 green creature
// that's still a land. (This effect lasts indefinitely.)" (Quirion Druid):
// a layer 4, 5 and 7b change pinned to the land object.
func S82OldB788AnimateTargetLand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	mods := []game.Mod{game.AddTypesMod("Creature"), game.SetColorsMod(ctx.TextColors("G")...)}
	mods = append(mods, game.SetBasePTMods(2, 2)...)
	return ScopedEffectFor{
		Target:   id,
		Mods:     mods,
		Duration: g.PinnedTo(game.IndefiniteDuration(), id),
		Label:    "Quirion Druid — becomes a 2/2 green creature that's still a land",
	}.Apply(ctx)
}

// S82OldB788IsToken is the card predicate for "token" (Diamond Kaleidoscope's
// "Sacrifice a Prism token").
func S82OldB788IsToken(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsToken() }

// S82OldB788CreaturesYouControlOfColorHaveProtection is "<Color> creatures
// you control have protection from <color>" (Righteous War): a Layer-6
// keyword grant over the source controller's creatures of one colour, read
// through the layered view.
func S82OldB788CreaturesYouControlOfColorHaveProtection(color, keyword, label string) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && target.HasColor(color)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Abilities = game.AppendKeywordAbility(c.Abilities, keyword)
		},
	}
}

// S82OldB788AnyTypeTheSacrificedLandCouldProduce is Squandered Resources'
// mana: "Add one mana of any type the sacrificed land could produce." The
// types were read off the land while it was still on the battlefield
// (game.PaidCost.SacrificedMana); the controller picks one of them, and a
// land that could produce nothing adds nothing.
func S82OldB788AnyTypeTheSacrificedLandCouldProduce(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	if len(paid.SacrificedMana) == 0 || paid.SacrificedMana[0] == "" {
		return game.ManaOutput{}
	}
	return game.ManaOutput{Text: "{" + strings.Join(strings.Split(paid.SacrificedMana[0], ""), "|") + "}"}
}
