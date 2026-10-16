package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sauron, the Necromancer — Legendary Creature — Avatar Horror
// {3}{B}{B}, 4/4:
//
//	"Menace
//	 Whenever Sauron attacks, exile target creature card from your
//	 graveyard. Create a tapped and attacking token that's a copy of
//	 that card, except it's a 3/3 black Wraith with menace. At the
//	 beginning of the next end step, exile that token unless Sauron is
//	 your Ring-bearer."
//
// The copy is of the card as printed (CR 707.2), read before it is
// exiled; the exception replaces its colors and its creature types and
// sets its base power and toughness (CR 707.9b), and those are the
// token's copiable values (2023-06-16 ruling). It was never declared
// as an attacker, so "whenever a creature attacks" doesn't trigger for
// it (CR 508.4).
//
// The end-step exile is a delayed trigger (CR 603.7) holding the token
// and Sauron as objects. As it resolves, the token stays only if that
// same Sauron is on the battlefield as its controller's Ring-bearer; a
// Sauron that left, even one that came back, does not save it (ruling).
//
// CR 508.4 lets the token's controller choose which defending player,
// planeswalker or battle it attacks, and CreateTokensAttackingYourChoice
// asks (a bot sends it where Sauron went). No simplification.
func init() {
	sauronNecromancerExileBody = game.DelayedBody("sauron-the-necromancer/exile-token-unless-ring-bearer", sauronNecromancerExileToken)
	attacks := WheneverThisAttacks("Sauron, the Necromancer — exile a creature card and make an attacking Wraith copy", sauronNecromancerRaise)
	attacks.Targets = TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())
	Register(Spec{
		OracleID:        "8dcd7e2b-75a9-476a-a0a0-798613e8dad6",
		Name:            "Sauron, the Necromancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{attacks},
	})
}

var sauronNecromancerExileBody game.BodyRef

// sauronNecromancerRaise is the attack trigger's body.
func sauronNecromancerRaise(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(g, ids[0])
	if !ok {
		return nil
	}
	sauronWraithExceptionFor(ctx.textReader())(&tmpl)
	prefer := triggerAttackTarget(item)
	sauron := game.ObjectRef{ID: item.SourceCardID, Epoch: -1}
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		sauron.Epoch = c.ObjectEpoch
	}
	return ExileTarget{
		Target: ids[0],
		Then: func(ctx *Context, _ bool) error {
			item := ctx.Item
			return CreateTokensAttackingYourChoice{
				Controller: item.Controller,
				Template:   tmpl,
				N:          1,
				Tapped:     true,
				Prefer:     prefer,
				Label:      "Sauron, the Necromancer",
				Then: func(ctx *Context, created []uuid.UUID) error {
					if len(created) == 0 {
						return nil
					}
					return ScheduleDelayedTrigger{
						Label:  "Sauron, the Necromancer — exile the Wraith unless Sauron is your Ring-bearer",
						Cards:  created,
						Body:   sauronNecromancerExileBody,
						Params: game.EffectParams{Object: sauron},
					}.Apply(ctx)
				},
			}.Apply(NewContext(ctx.Game, item))
		},
	}.Apply(ctx)
}

// sauronWraithException is "except it's a 3/3 black Wraith with
// menace", on a template that is also tapped: the colors and the
// creature types are replaced, supertypes and card types stay. Read in
// the printed words; the trigger reads them through Sauron's text
// (sauronWraithExceptionFor).
func sauronWraithException(t *game.Card) { sauronWraithExceptionFor(nil)(t) }

// sauronWraithExceptionFor is the exception with "black" and "Wraith"
// read through `reader`, the object whose text it is (CR 612.2, #3984);
// nil reads the printed words.
func sauronWraithExceptionFor(reader *game.Card) func(*game.Card) {
	return func(t *game.Card) {
		t.Power, t.Toughness = 3, 3
		t.VariableToughness = false // a printed 3, not a `*` stand-in (#683)
		t.SetCopyExceptionColors(reader.TextColor("B"))
		t.TypeLine = retypedTypeLine(t.TypeLine, reader.TextSubtype("Wraith"))
		if !slices.Contains(t.Keywords, "menace") {
			t.Keywords = append(t.Keywords, "menace")
		}
		t.Tapped = true
	}
}

// sauronNecromancerExileToken is the delayed trigger's body.
func sauronNecromancerExileToken(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if c, ok := g.LookupCardForEffect(p.Object.ID); ok && c.ObjectEpoch == p.Object.Epoch &&
		onBattlefield(g, p.Object.ID) && game.IsRingBearerOf(c, item.Controller) {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if !onBattlefield(g, t.ID) {
			continue
		}
		if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
