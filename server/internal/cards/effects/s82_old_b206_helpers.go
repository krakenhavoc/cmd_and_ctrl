package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b206_helpers.go — shared bodies for Battlebond, batch 206 (#5529).

// OldB206FriendFoe is the two halves of a "for each player, choose friend or
// foe" card (Battlebond): what each friend does and what each foe does. Each
// half queues its own prompt for that player and returns; the halves are
// plain code, looked up by the key the card registered them under, so the
// pending question carries data only.
type OldB206FriendFoe struct {
	Name   string
	Friend func(ctx *Context, who uuid.UUID) error
	Foe    func(ctx *Context, who uuid.UUID) error
}

var oldB206Kinds = map[string]OldB206FriendFoe{}

// RegisterOldB206FriendFoe records a card's two halves under its name; call it
// from the card's init.
func RegisterOldB206FriendFoe(k OldB206FriendFoe) { oldB206Kinds[k.Name] = k }

// OldB206ChooseFriendOrFoe is the resolution of the card: the caster is asked,
// one player at a time starting with themselves, whether that player is a
// friend or a foe; once every answer is in, each friend and then each foe acts
// in turn order from the caster. The answers ride the prompt as words
// ("<card>", then "F:<id>" / "O:<id>").
func OldB206ChooseFriendOrFoe(ctx *Context, name string) error {
	return oldB206Ask(ctx, name, SeatsFromController(ctx.Game, ctx.Controller()), []string{name})
}

func oldB206Ask(ctx *Context, name string, seats []uuid.UUID, answers []string) error {
	for len(seats) > 0 {
		p := ctx.Game.PlayerByIDForEffect(seats[0])
		if p == nil || p.Eliminated {
			seats = seats[1:]
			continue
		}
		return PickOption{
			Question: name + " — is " + p.Name + " a friend or a foe?",
			Options:  []game.ChoiceOption{{Label: "Friend"}, {Label: "Foe"}},
			ThenKey:  oldB206Answer,
			Carry:    seats,
			Words:    answers,
			WithItem: true,
		}.Apply(ctx)
	}
	return oldB206Act(ctx, answers)
}

// oldB206Answer records the answer for the first seat of the carry and asks
// about the next. An on-disk key.
var oldB206Answer game.OptionPickThen

// oldB206AnswerRegistered registers the key at package initialisation. It is
// a value rather than an init func, and the answer is read through
// oldB206Answer, which has no initialiser, so the question and its answer may
// call each other.
var _ = func() bool {
	oldB206Answer = OptionPickThen("option-pick/s82-old-b206-friend-or-foe", oldB206Answered)
	return true
}()

func oldB206Answered(ctx *Context, r game.OptionPicked) error {
	answers := append([]string(nil), r.Words...)
	if len(answers) == 0 || len(r.Carry) == 0 {
		return oldB206Act(ctx, answers)
	}
	who, rest := r.Carry[0], r.Carry[1:]
	switch r.Index {
	case 0:
		answers = append(answers, "F:"+who.String())
	case 1:
		answers = append(answers, "O:"+who.String())
	}
	return oldB206Ask(ctx, answers[0], rest, answers)
}

// oldB206Act runs each friend's half, then each foe's, in the order the
// caster answered.
func oldB206Act(ctx *Context, answers []string) error {
	if len(answers) == 0 {
		return nil
	}
	k, ok := oldB206Kinds[answers[0]]
	if !ok {
		return nil
	}
	for _, want := range []byte{'F', 'O'} {
		for _, a := range answers[1:] {
			if len(a) < 3 || a[0] != want {
				continue
			}
			who, err := uuid.Parse(a[2:])
			if err != nil {
				continue
			}
			p := ctx.Game.PlayerByIDForEffect(who)
			if p == nil || p.Eliminated {
				continue
			}
			half := k.Foe
			if want == 'F' {
				half = k.Friend
			}
			if err := half(ctx, who); err != nil {
				return err
			}
		}
	}
	return nil
}

// OldB206PickToSacrifice asks who to sacrifice one permanent they control that
// matches; nobody is asked when nothing matches.
func OldB206PickToSacrifice(ctx *Context, who uuid.UUID, question string, match CardPredicate) {
	var ids []uuid.UUID
	for _, c := range ctx.Game.Battlefield.Cards {
		if c.Controller == who && match(ctx.Game, who, c) {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  who,
		Source:   ctx.Source(),
		Question: question,
		Cards:    ids,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  oldB206Sacrificed,
	})
}

var oldB206Sacrificed = game.RegisterChooseCardsThen("choose-cards/s82-old-b206-sacrifice-picked", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(r.Picked[0]); ok && g.Battlefield.Contains(c.InstanceID) && c.Controller == r.Chooser {
		return g.SacrificePermanentForEffect(c.InstanceID)
	}
	return nil
})

// OldB206PickGraveyardToHand asks who to return a card matching from their own
// graveyard to their hand; nobody is asked when nothing matches.
func OldB206PickGraveyardToHand(ctx *Context, who uuid.UUID, question string, match func(game.Card) bool) {
	ids := GraveyardCardIDs(ctx, who, match)
	if len(ids) == 0 {
		return
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  who,
		Source:   ctx.Source(),
		Question: question,
		Cards:    ids,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  oldB206Returned,
	})
}

var oldB206Returned = game.RegisterChooseCardsThen("choose-cards/s82-old-b206-return-picked", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	ctx := NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: r.Chooser, Owner: r.Chooser, SourceCardID: r.Source})
	return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}.Apply(ctx)
})

// OldB206AnyPlayerCastsCondition is "whenever a player casts" a spell passing
// the filter in the trigger's params: any caster, not only the trigger's
// controller (YouNextCast's twin without the "you").
var OldB206AnyPlayerCastsCondition = game.DelayedCondition("cast/any-player-casts",
	func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
		if ev.Kind != game.EventCast {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return false
		}
		x := 0
		if item := g.StackItemForEffect(ev.CardID); item != nil && item.XValue > 0 {
			x = item.XValue
		}
		return p.Filter.MatchesSpell(c, x)
	})

// OldB206CopyTheSpellForItsCasterBody is "that player copies it and may choose
// new targets for the copy": the copy is created by the spell's caster, who
// controls it and chooses (CR 707.10), read from the event that fired the
// trigger so it is right when the caster is not the card's controller.
var OldB206CopyTheSpellForItsCasterBody = game.SimpleDelayedBody("copy/the-spell-for-its-caster", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cast := ctx.PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	caster := item.Controller
	if s := g.StackItemForEffect(cast[0]); s != nil {
		caster = s.Controller
	}
	return CopySpell{StackID: cast[0], Controller: caster, ChooseNewTargets: true, FromLastKnown: true}.Apply(ctx)
})

// OldB206InstantOrSorceryCastOnYourTurn is "whenever an instant or sorcery
// spell is cast during your turn": any player's cast, while the source's
// controller is the active player.
func OldB206InstantOrSorceryCastOnYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || !IsYourTurn(g, source.Controller) {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}

// OldB206InstantsAndSorceriesCastBefore counts the instant and sorcery spells
// cast this turn before the spell the event names, by any player, off the
// event log back to the turn's start. A copy is created, not cast (CR 707.10),
// so it is not counted.
func OldB206InstantsAndSorceriesCastBefore(g *game.Game, ev game.Event) int {
	n := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		prior := g.Events[i]
		if prior.Kind == game.EventTurnBegan {
			break
		}
		if prior.Seq >= ev.Seq || prior.Kind != game.EventCast {
			continue
		}
		if c, ok := g.LookupCardForEffect(prior.CardID); ok && (c.IsInstant() || c.IsSorcery()) {
			n++
		}
	}
	return n
}

// OldB206FumbleResolve is Fumble's resolution: return the target creature to
// its owner's hand, gain control of the Auras and Equipment that were attached
// to it, then (when there are any) the controller chooses another creature and
// each of them that can legally be attached to it is. The attachments are read
// before the creature leaves, since they fall off with it, and ride the bounce's
// tail, which runs once the creature has really gone. An attachment the chosen
// creature cannot take stays unattached, where the usual state-based actions
// find it (CR 704.5m).
func OldB206FumbleResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	var attached []uuid.UUID
	for _, aid := range g.AttachmentsOf(id) {
		if a, found := g.LookupCardForEffect(aid); found && (a.IsAura() || a.HasSubtype("Equipment")) {
			attached = append(attached, aid)
		}
	}
	tail := ctx.ItemTail(oldB206FumbleTail)
	tail.Carry = append([]uuid.UUID{id}, attached...)
	return BounceToHand{Target: id, ThenTail: tail}.Apply(ctx)
}

// oldB206FumbleTail is the rest of Fumble once the creature has been returned:
// gain control of each Aura and Equipment that was attached to it, then offer
// the choice of another creature. An on-disk key: never renamed, never reused.
var oldB206FumbleTail = CardTail("then/s82-old-b206-fumble-gain-and-attach", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	target := o.Tail.Carry[0]
	var kept []uuid.UUID
	for _, aid := range o.Tail.Carry[1:] {
		if !g.Battlefield.Contains(aid) {
			continue
		}
		if err := (GainControl{Target: aid, Controller: ctx.Controller(), Label: "Fumble — gain control"}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		kept = append(kept, aid)
	}
	if len(kept) == 0 {
		return nil
	}
	var hosts []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.InstanceID != target {
			hosts = append(hosts, c.InstanceID)
		}
	}
	if len(hosts) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Fumble — attach the Auras and Equipment to another creature",
		Cards:    hosts,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  oldB206FumbleAttach,
		Carry:    kept,
	})
	return nil
})

// oldB206FumbleAttach attaches each carried Auras and Equipment that can
// legally be attached to the chosen creature. An on-disk key: never renamed,
// never reused.
var oldB206FumbleAttach = game.RegisterChooseCardsThen("choose-cards/s82-old-b206-fumble-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	host := r.Picked[0]
	for _, aid := range r.Carry {
		if !g.AttachLegalForEffect(aid, host) {
			continue
		}
		if err := g.AttachForEffect(aid, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			return err
		}
	}
	return nil
})
