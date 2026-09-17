package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armored Skyhunter — Creature — Cat Knight {3}{W}, 3/3:
//
//	"Flying
//	 Whenever this creature attacks, look at the top six cards of your
//	 library. You may put an Aura or Equipment card from among them onto
//	 the battlefield. If an Equipment is put onto the battlefield this
//	 way, you may attach it to a creature you control. Put the rest of
//	 those cards on the bottom of your library in a random order."
//
// The attack trigger is #745's PutFromLibraryOntoBattlefield over the
// six cards the controller looked at (nobody else sees them). The rest
// go to the bottom in an order drawn from the game's seeded RNG, and an
// Equipment that entered is offered to a creature. The attach is part
// of the same resolution — not a reflexive trigger — so it is asked
// inline: a choose_cards prompt over the creatures the controller has,
// with "none" a legal answer. The bottoming does not wait for that
// answer, which nothing about the answer could change.
//
// DECLARED SIMPLIFICATION, weaker than printed: Aura cards are not
// offered, only Equipment. An Aura put onto the battlefield without
// being cast needs its controller to choose what it enchants as it
// enters (CR 303.4f), and the library-to-battlefield move has no entry
// prompt to ask that with — the same reason every library, hand and
// search entry skips the Clone choice. Offering an Aura it then could
// not attach would put a permanent onto the battlefield that the next
// state-based check sends to the graveyard.
func init() {
	Register(Spec{
		OracleID:        "b4dbbf56-d7df-4183-bb48-cfc6b4d0468f",
		Name:            "Armored Skyhunter",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"Only an Equipment card can be put onto the battlefield — Aura cards among the six aren't offered."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Armored Skyhunter — look at the top six cards; you may put an Equipment onto the battlefield",
				armoredSkyhunterLook),
		},
	})
}

func armoredSkyhunterLook(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	return PutFromLibraryOntoBattlefield{
		Player:   controller,
		Cards:    g.LookAtTopOfLibraryForEffect(controller, 6),
		Match:    OfSubtype("Equipment"),
		Max:      1,
		Optional: true,
		Label:    "Armored Skyhunter — you may put an Equipment card from among them onto the battlefield",
		Then: func(g *game.Game, res PutFromLibraryResult) error {
			if err := PutRestOnBottomInRandomOrder(g, res); err != nil {
				return err
			}
			if len(res.Entered) == 0 {
				return nil
			}
			return offerAttachToACreatureYouControl(g, source, controller, res.Entered[0],
				"Armored Skyhunter — you may attach the Equipment to a creature you control")
		},
	}.Apply(NewContext(g, item))
}

// offerAttachToACreatureYouControl is "you may attach it to a creature
// you control", asked as part of a resolution: a choose_cards prompt
// over the controller's creatures with a floor of zero. The pick is
// re-checked on submit (it must still be on the battlefield) and the
// attach refuses a host that has since left.
//
// Caller holds g.mu.
func offerAttachToACreatureYouControl(g *game.Game, source, controller, attachment uuid.UUID, question string) error {
	var creatures []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.InstanceID != attachment && c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: question,
		Cards:    creatures,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			host, ok := g.LookupCardForEffect(picked[0])
			if !ok || host.Controller != controller {
				return nil
			}
			return g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: picked[0]})
		},
	})
	return nil
}
