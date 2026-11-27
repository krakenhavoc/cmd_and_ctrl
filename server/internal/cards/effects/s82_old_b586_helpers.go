package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b586_helpers.go — Saviors of Kamigawa, batch 586 (#5601): Choice of
// Damnations' number and edict, Blood Clock's upkeep question and Measure of
// Wickedness's hand-over.

// OldB586YouDontControlA is "you don't control a <subtype>", read through the
// layers (Raving Oni-Slave's Demon).
func OldB586YouDontControlA(subtype string) func(g *game.Game, you uuid.UUID) bool {
	controls := ControlsA(subtype)
	return func(g *game.Game, you uuid.UUID) bool { return !controls(g, you) }
}

// OldB586YouHaveSevenCardsInHand is "you have seven or more cards in hand".
var OldB586YouHaveSevenCardsInHand = YouHaveCardsInHandAtLeast(7)

// OldB586GainLifeEqualToThatSpellsManaValue is Bounteous Kirin's gain: the cast
// spell's mana value, read as the trigger resolves (with its announced X while
// it is still on the stack).
func OldB586GainLifeEqualToThatSpellsManaValue(g *game.Game, item *game.StackItem) error {
	n := TriggeringSpellManaValue(g, item)
	if n <= 0 {
		return nil
	}
	return GainLife{Amount: n}.Apply(NewContext(g, item))
}

// OldB586AnotherCardPutIntoYourGraveyard matches "another card is put into
// your graveyard from anywhere": a card (not a token) its controller owns
// arriving in the graveyard from any other zone, this permanent excepted.
func OldB586AnotherCardPutIntoYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken() && c.Owner == source.Controller
}

// OldB586SacrificeThisAndLose is "sacrifice this enchantment and you lose N
// life": the loss does not depend on the sacrifice happening.
func OldB586SacrificeThisAndLose(n int) Effect {
	return Do(
		Instruction(func(ctx *Context) error { return SacrificeThisIfStillOnBattlefield(ctx.Game, ctx.Item) }),
		YouLoseLife(n),
	)
}

// OldB586TargetOpponentGainsControlOfThis is "target opponent gains control of
// this <permanent>", indefinitely, while it is still the permanent the ability
// came from (CR 400.7).
func OldB586TargetOpponentGainsControlOfThis(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 || !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		return GainControl{
			Target:     item.SourceCardID,
			Controller: targets[0].ID,
			Duration:   game.IndefiniteDuration(),
			Label:      label,
		}.Apply(ctx)
	}
}

// --- Choice of Damnations ----------------------------------------------

// OldB586ChoiceOfDamnations asks the target opponent for a number; the
// caster then decides between the life loss and the sacrifice.
func OldB586ChoiceOfDamnations(item *game.StackItem, ctx *Context) error {
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetPlayer {
		return nil
	}
	opp := targets[0].ID
	return ChooseNumber{
		Chooser:  opp,
		Min:      0,
		NoMax:    true,
		Question: "Choice of Damnations — choose a number",
		Unit:     game.PayAmountOther,
		Then:     oldB586DamnationsNumber,
		Carry:    []uuid.UUID{opp},
	}.Apply(ctx)
}

// oldB586DamnationsNumber hands the chosen number to the caster: they may have
// the opponent lose that much life. An on-disk key.
var oldB586DamnationsNumber = NumberThen("number/old-b586-choice-of-damnations", func(ctx *Context, r game.NumberAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	opp := r.Carry[0]
	if p := ctx.Game.PlayerByIDForEffect(opp); p == nil || p.Eliminated {
		return nil
	}
	return MayChoice{
		Player:   ctx.Controller(),
		Question: "Choice of Damnations — have that player lose " + strconv.Itoa(r.Number) + " life? If you don't, they sacrifice all but " + strconv.Itoa(r.Number) + " permanents.",
		Then:     oldB586DamnationsYesNo,
		Carry:    []uuid.UUID{opp},
		Word:     strconv.Itoa(r.Number),
	}.Apply(ctx)
})

// oldB586DamnationsYesNo is the caster's answer: yes, the opponent loses that
// much life; no, they sacrifice all but that many permanents of their choice.
var oldB586DamnationsYesNo = MayYesNo("confirm/old-b586-choice-of-damnations",
	func(ctx *Context, r game.ConfirmAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Carry[0], -WordInt(r))
	},
	func(ctx *Context, r game.ConfirmAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return oldB586SacrificeAllBut(ctx, r.Carry[0], WordInt(r))
	})

// oldB586SacrificeAllBut makes `player` keep `keep` permanents of their choice
// and sacrifice the rest, all at once (CR 701.21). With `keep` or more
// permanents nothing goes; with none to keep, everything does.
func oldB586SacrificeAllBut(ctx *Context, player uuid.UUID, keep int) error {
	all := PermanentsControlledBy(ctx.Game, player)
	switch {
	case len(all) == 0 || keep >= len(all):
		return nil
	case keep <= 0:
		ctx.Game.SacrificeAllForEffect(ctx.Source(), all)
		return nil
	}
	return ChoosePermanents{
		Player:   player,
		Question: "Choice of Damnations — choose " + strconv.Itoa(keep) + " permanents to keep; sacrifice the rest",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return PermanentsControlledBy(g, of), keep, keep
		},
		Of:    []uuid.UUID{player},
		Keyed: oldB586SacrificeTheRest,
	}.Apply(ctx)
}

// oldB586SacrificeTheRest sacrifices everything the chooser controls that they
// did not pick. An on-disk key.
var oldB586SacrificeTheRest = RegisterPermanentsPick("prompt-run/old-b586-sacrifice-all-but", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		kept := map[uuid.UUID]bool{}
		for _, id := range picked.Cards() {
			kept[id] = true
		}
		for _, sc := range picked {
			var rest []uuid.UUID
			for _, id := range PermanentsControlledBy(ctx.Game, sc.Seat) {
				if !kept[id] {
					rest = append(rest, id)
				}
			}
			ctx.Game.SacrificeAllForEffect(ctx.Source(), rest)
		}
		return nil
	},
})

// --- Blood Clock --------------------------------------------------------

const (
	oldB586ClockReturn = "Return a permanent you control to its owner's hand"
	oldB586ClockPay    = "Pay 2 life"
)

// OldB586BloodClock is the upkeep trigger: the player whose upkeep it is
// returns a permanent they control to its owner's hand unless they pay 2 life.
// With no permanent to return there is nothing to ask; with under 2 life they
// cannot pay (CR 119.4) and so return one.
func OldB586BloodClock(g *game.Game, item *game.StackItem) error {
	player := item.Trigger.Event.Actor
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Eliminated || len(PermanentsControlledBy(g, player)) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	if p.Life < 2 {
		return oldB586ClockReturnAPermanent(ctx, player)
	}
	return PickOption{
		Player:   player,
		Question: "Blood Clock — return a permanent you control to its owner's hand unless you pay 2 life",
		Options: []game.ChoiceOption{
			{Label: oldB586ClockReturn},
			{Label: oldB586ClockPay, LifeCost: 2},
		},
		ThenKey: oldB586ClockAnswer,
	}.Apply(ctx)
}

var oldB586ClockAnswer = OptionPickThen("option-pick/old-b586-blood-clock", func(ctx *Context, r game.OptionPicked) error {
	if r.Option != nil && r.Option.Label == oldB586ClockPay {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -2)
	}
	return oldB586ClockReturnAPermanent(ctx, r.Chooser)
})

func oldB586ClockReturnAPermanent(ctx *Context, player uuid.UUID) error {
	return ChoosePermanents{
		Player:   player,
		Question: "Blood Clock — return a permanent you control to its owner's hand",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return PermanentsControlledBy(g, of), 1, 1
		},
		Of:    []uuid.UUID{player},
		Keyed: oldB586ClockBounced,
	}.Apply(ctx)
}

var oldB586ClockBounced = RegisterPermanentsPick("prompt-run/old-b586-blood-clock-return", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
