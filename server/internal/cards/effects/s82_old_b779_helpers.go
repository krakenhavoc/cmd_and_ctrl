package effects

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b779_helpers.go — S82 OLD batch 779 (Weatherlight, #5629).

// S82OldB779CreaturesCantAttack is Peacekeeper's static, "Creatures can't
// attack" (CR 508.1c): every creature on the battlefield, whoever controls it,
// carries the can't-attack restriction.
func S82OldB779CreaturesCantAttack(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsCreature() },
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= game.CantAttack
		},
	}
}

// S82OldB779UntappedPermanentsCount is Maraxus of Keld's number: the untapped
// artifacts, creatures and lands `controller` controls, each permanent counted
// once however many of the three it is. It walks the live battlefield because it
// runs inside a layer recompute (as Lumra's lands count does).
func S82OldB779UntappedPermanentsCount(g *game.Game, controller uuid.UUID) int {
	if g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && !c.Tapped && (c.IsArtifact() || c.IsCreature() || c.IsLand()) {
			n++
		}
	}
	return n
}

// s82OldB779GainControlOfBlockerBody is the delayed trigger of "gain control of
// that creature at end of combat" (Tolarian Entrancer): the blocker named by the
// trigger's object, if it is still that object on the battlefield (CR 400.7),
// comes under the delayed trigger's controller's control with no duration
// (CR 611.2). A registered key (ADR 0041 phase 3).
var s82OldB779GainControlOfBlockerBody = game.SimpleDelayedBody("s82-old-b779/gain-control-of-blocker", func(g *game.Game, item *game.StackItem) error {
	obj := item.Params.Object
	if obj.ID == uuid.Nil || !OnBattlefield(g, obj.ID) || g.ObjectEpochForEffect(obj.ID) != obj.Epoch {
		return nil
	}
	return GainControl{Target: obj.ID, Controller: item.Controller, Duration: game.IndefiniteDuration(),
		Label: "Tolarian Entrancer — gain control of the creature that blocked it"}.Apply(NewContext(g, item))
})

// S82OldB779GainControlOfTheBlockerAtEndOfCombat is Tolarian Entrancer's body:
// the creature the event says blocked it comes under its controller's control
// at the beginning of this combat's end of combat step (CR 511.2).
func S82OldB779GainControlOfTheBlockerAtEndOfCombat(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
			return nil
		}
		blocker := item.Trigger.Event.CardID
		if !OnBattlefield(g, blocker) {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:     game.StepEndCombat,
			Label:  label,
			Params: game.EffectParams{Object: game.ObjectRef{ID: blocker, Epoch: g.ObjectEpochForEffect(blocker)}},
			Body:   s82OldB779GainControlOfBlockerBody,
		}.Apply(NewContext(g, item))
	}
}

// S82OldB779CreaturesPutIntoYourGraveyardThisTurn is Urborg Justice's count:
// the creatures put into `owner`'s graveyard from the battlefield this turn,
// tokens included (a token is put into the graveyard before it ceases to exist,
// CR 111.7). A card is put into its owner's graveyard (CR 400.3); a token that
// has already gone is read by who controlled it as it left.
func S82OldB779CreaturesPutIntoYourGraveyardThisTurn(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok {
			if LeftAsType(ev, c, "creature") && c.Owner == owner {
				n++
			}
			continue
		}
		if was, known := ev.WasType("creature"); known && was {
			if controller, ok := ev.LeftUnderControlOf(); ok && controller == owner {
				n++
			}
		}
	}
	return n
}

// S82OldB779UrborgJusticeSacrifices is Urborg Justice's resolution: the target
// opponent sacrifices a creature of their choice for each creature put into the
// caster's graveyard from the battlefield this turn, one after another (CR 701.21).
func S82OldB779UrborgJusticeSacrifices(ctx *Context) error {
	opp := TargetedPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	n := S82OldB779CreaturesPutIntoYourGraveyardThisTurn(ctx.Game, ctx.Controller())
	if n == 0 {
		return nil
	}
	ctx.Game.PlayerSacrificesNForEffect(ctx.Source(), opp, SacrificeSpec("a creature", Creature()), "Urborg Justice — sacrifice a creature", n)
	return nil
}

// S82OldB779CreaturesThatBlockedOrWereBlockedThisTurn is Heat Stroke's set:
// every creature that blocked, or was blocked, this turn (one EventBlock per
// pair: CardID the blocker, Target the attacker), by objects, not names
// (CR 400.7): a creature that has since changed zones is no longer the one that
// blocked.
func S82OldB779CreaturesThatBlockedOrWereBlockedThisTurn(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for _, ev := range g.EventsThisTurn() {
		switch ev.Kind {
		case game.EventZoneMove:
			if i := slices.Index(out, ev.CardID); i >= 0 {
				out = slices.Delete(out, i, i+1)
			}
		case game.EventBlock:
			for _, id := range []uuid.UUID{ev.CardID, ev.Target} {
				if id != uuid.Nil && !slices.Contains(out, id) {
					out = append(out, id)
				}
			}
		}
	}
	return out
}

// S82OldB779DestroyCreaturesThatBlockedOrWereBlocked is Heat Stroke's body at
// each end of combat: the creatures of that set still on the battlefield are
// destroyed together.
func S82OldB779DestroyCreaturesThatBlockedOrWereBlocked(g *game.Game, item *game.StackItem) error {
	set := map[uuid.UUID]bool{}
	for _, id := range S82OldB779CreaturesThatBlockedOrWereBlockedThisTurn(g) {
		set[id] = true
	}
	if len(set) == 0 {
		return nil
	}
	return DestroyAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return set[c.InstanceID] && c.IsCreature() },
	}.Apply(NewContext(g, item))
}

// S82OldB779Tariff is Tariff's resolution. In turn order from the active
// player each player picks the creature they control with the greatest mana
// value (a player with two or more tied picks one of them, a player with one is
// not asked), and then sacrifices it unless they pay its mana cost. A creature
// with no mana cost has nothing that can be paid, so it is sacrificed at once
// (CR 118.3).
func S82OldB779Tariff(ctx *Context) error {
	g := ctx.Game
	for _, player := range TablePlayers(ctx) {
		tied := s82OldB779GreatestManaValueCreatures(g, player)
		switch len(tied) {
		case 0:
		case 1:
			if err := s82OldB779TariffPay(ctx, player, tied[0]); err != nil {
				return err
			}
		default:
			if err := (ChoosePermanents{
				Player:   player,
				Of:       []uuid.UUID{player},
				Question: "Tariff — choose the creature you control with the greatest mana value",
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					return s82OldB779GreatestManaValueCreatures(g, of), 1, 1
				},
				Keyed: s82OldB779TariffPicked,
			}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// s82OldB779GreatestManaValueCreatures is the creatures `player` controls that
// tie for the greatest mana value among them.
func s82OldB779GreatestManaValueCreatures(g *game.Game, player uuid.UUID) []uuid.UUID {
	best := -1
	var out []uuid.UUID
	for _, id := range PermanentsControlledByMatching(g, player, Creature()) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		switch mv := c.ManaValue(); {
		case mv > best:
			best, out = mv, []uuid.UUID{id}
		case mv == best:
			out = append(out, id)
		}
	}
	return out
}

// s82OldB779TariffPicked is a tied player's pick answered: the chosen creature
// is paid for or sacrificed. An on-disk key: never renamed, never reused.
var s82OldB779TariffPicked = RegisterPermanentsPick("prompt-run/s82-old-b779-tariff-picked", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				continue
			}
			if err := s82OldB779TariffPay(ctx, c.Controller, id); err != nil {
				return err
			}
		}
		return nil
	},
})

// s82OldB779TariffPay asks `player` to pay the creature's mana cost or have it
// sacrificed.
func s82OldB779TariffPay(ctx *Context, player, creature uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(creature)
	if !ok || !OnBattlefield(ctx.Game, creature) {
		return nil
	}
	cost := strings.ReplaceAll(c.ManaCost, "{X}", "")
	if strings.TrimSpace(cost) == "" {
		return SacrificePermanent{Target: creature}.Apply(ctx)
	}
	return PayUnless{
		Chooser:  player,
		Cost:     cost,
		Question: "Tariff — pay " + cost + " or sacrifice " + c.Name + "?",
		Else:     s82OldB779TariffDeclined,
		Carry:    []uuid.UUID{creature},
	}.Apply(ctx)
}

// s82OldB779TariffDeclined is the "unless they pay" not paid: the creature is
// sacrificed if it is still on the battlefield. An on-disk key.
var s82OldB779TariffDeclined = PayUnlessElse("pay-unless/s82-old-b779-tariff-sacrifice", func(ctx *Context, r game.PayUnlessAnswer) error {
	if len(r.Carry) == 0 || !OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	return SacrificePermanent{Target: r.Carry[0]}.Apply(ctx)
})

// S82OldB779ParadigmShift is Paradigm Shift's resolution: the controller's
// whole library is exiled, then their graveyard is shuffled into their library
// (CR 608.2c's order).
func S82OldB779ParadigmShift(ctx *Context) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Library == nil {
		return nil
	}
	var library []uuid.UUID
	for _, c := range p.Library.Cards {
		library = append(library, c.InstanceID)
	}
	return ctx.Game.ExileCardsKeyedForEffect(library, ctx.Tail(s82OldB779ShuffleGraveyardIn))
}

// s82OldB779ShuffleGraveyardIn is "then shuffle your graveyard into your
// library", once the library has been exiled. An on-disk key.
var s82OldB779ShuffleGraveyardIn = CardTail("s82-old-b779/shuffle-graveyard-into-library", func(ctx *Context, _ game.TailOutcome) error {
	player := ctx.Controller()
	if p := ctx.Game.PlayerByIDForEffect(player); p != nil && p.Graveyard != nil {
		var ids []uuid.UUID
		for _, c := range p.Graveyard.Cards {
			ids = append(ids, c.InstanceID)
		}
		for _, id := range ids {
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return ShuffleLibrary{Player: player}.Apply(ctx)
})

// S82OldB779DestroyTargetsTail is Goblin Grenadiers' "if you do, destroy target
// creature and target land": once the Grenadiers really was sacrificed, each of
// the trigger's two targets that is still legal is destroyed (CR 608.2b). A
// registered key (ADR 0041 phase 3).
var S82OldB779DestroyTargetsTail = CardTail("s82-old-b779/destroy-the-two-targets", func(ctx *Context, o game.TailOutcome) error {
	if !OneWentThisWay(o) {
		return nil
	}
	for slot := 0; slot < 2; slot++ {
		if t, ok := ctx.ClauseTarget(slot); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
})
