package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gilgamesh, Master-at-Arms — Legendary Creature — Human Samurai
// {4}{R}{R}, 6/6:
//
//	"Whenever Gilgamesh enters or attacks, look at the top six cards of
//	 your library. You may put any number of Equipment cards from among
//	 them onto the battlefield. Put the rest on the bottom of your
//	 library in a random order. When you put one or more Equipment onto
//	 the battlefield this way, you may attach one of them to a Samurai
//	 you control."
//
// The shared "look at the top N, put some onto the battlefield" sentence
// (#745) with "any number of Equipment cards": the six cards are looked
// at, only Gilgamesh's controller sees them, the chosen Equipment enter
// as one simultaneous batch without being cast, and the rest go to the
// bottom in an order drawn from the game's seeded RNG.
//
// The reflexive sentence (CR 603.12) is a real ReflexiveTrigger: once at
// least one Equipment entered, it goes on the stack above Gilgamesh's
// ability so the table can respond, carries the Equipment that entered
// as its payload, and on resolution asks which one (or none) goes onto
// which Samurai. Nothing in it targets, so both picks happen as it
// resolves, and an Equipment or Samurai that has left by then is skipped.
func init() {
	Register(Spec{
		OracleID:     "c9432e87-38f6-4889-8f07-ae7cd045e532",
		Name:         "Gilgamesh, Master-at-Arms",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Gilgamesh, Master-at-Arms — look at the top six cards; you may put any number of Equipment onto the battlefield",
				gilgameshLookAndPut),
		},
	})
}

// gilgameshLookAndPut looks at the top six, puts any number of Equipment
// onto the battlefield and the rest on the bottom in a random order, and
// then, if any Equipment entered, creates the reflexive attach trigger.
func gilgameshLookAndPut(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 6)
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    looked,
		Match:    OfSubtype("Equipment"),
		Max:      0,
		Optional: true,
		Label:    "Gilgamesh — put any number of Equipment cards from among them onto the battlefield",
		ThenTail: gilgameshRest,
		ThenItem: true,
	}.Apply(ctx)
}

// gilgameshRest is the rider as a keyed tail (#4253 slice 6): the rest to
// the bottom, then the reflexive attach trigger over the Equipment that
// entered. The asking item rides the tail. An on-disk key.
var gilgameshRest = PutFromLibraryTail("gilgamesh-master-at-arms/rest-then-attach-trigger",
	func(ctx *Context, res PutFromLibraryResult) error {
		g := ctx.Game
		if err := PutRestOnBottomInRandomOrder(g, res); err != nil {
			return err
		}
		if len(res.Entered) == 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label: "Gilgamesh, Master-at-Arms — you may attach one of the Equipment to a Samurai you control",
			Cards: res.Entered,
			Body:  gilgameshAttachBody,
		}.Apply(ctx)
	})

// gilgameshAttach is the reflexive half: pick one of the Equipment that
// entered (a floor of zero is the "may"), then the Samurai it goes on.
func gilgameshAttach(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	// "Samurai" is a word of the source's text (CR 612.2, #3984).
	word := g.TextReaderForEffect(item).TextSubtype("Samurai")
	var equipment []uuid.UUID
	for _, id := range NewContext(g, item).PayloadCards() {
		if c, ok := g.LookupCardForEffect(id); ok && c.HasSubtype("Equipment") && g.Battlefield.Contains(id) {
			equipment = append(equipment, id)
		}
	}
	if len(equipment) == 0 || len(gilgameshSamurai(g, controller, word)) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   item.SourceCardID,
		Question: "Gilgamesh — attach one of the Equipment to a Samurai you control?",
		Cards:    equipment,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			attachment := picked[0]
			samurai := gilgameshSamurai(g, controller, word)
			if len(samurai) == 0 {
				return nil
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  controller,
				Source:   item.SourceCardID,
				Question: "Gilgamesh — attach it to which Samurai you control?",
				Cards:    samurai,
				Min:      1,
				Max:      1,
				Zone:     game.ZoneBattlefield,
				Then: func(g *game.Game, hosts []uuid.UUID) error {
					if len(hosts) == 0 || !g.Battlefield.Contains(attachment) {
						return nil
					}
					host, ok := g.LookupCardForEffect(hosts[0])
					if !ok || host.Controller != controller || !host.HasSubtype(word) {
						return nil
					}
					return g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: hosts[0]})
				},
			})
			return nil
		},
	})
	return nil
}

// gilgameshSamurai lists the Samurai creatures a player controls,
// `samurai` being the creature type as the source's text reads it.
func gilgameshSamurai(g *game.Game, controller uuid.UUID, samurai string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsCreature() && c.HasSubtype(samurai) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
