package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Danitha, Benalia's Hope — Legendary Creature — Human Knight {4}{W}, 4/4
// (Dominaria United):
//
//	"First strike, vigilance, lifelink
//	 When Danitha enters, you may put an Aura or Equipment card from your
//	 hand or graveyard onto the battlefield attached to Danitha."
//
// Liberated Livestock's offer for a single host: when Danitha enters, a
// choose_cards prompt over the Aura and Equipment cards in the controller's
// hand and graveyard, with none a legal answer (offered only when there is
// a card to put, and not at all if Danitha has already left). The pick goes
// onto the battlefield attached to Danitha (CR 303.4f: the effect names the
// host, so an Aura is not asked for one); an Aura that can't legally enchant
// Danitha stays where it was (CR 303.4i).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "43b5dbe9-41f7-4082-944b-ea9d6292c6fd",
		Name:            "Danitha, Benalia's Hope",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Danitha, Benalia's Hope — you may put an Aura or Equipment card from your hand or graveyard onto the battlefield attached to Danitha",
				danithaOffer),
		},
	})
}

func danithaOffer(g *game.Game, item *game.StackItem) error {
	danitha := item.SourceCardID
	if !onBattlefield(g, danitha) || sourceIsNewObject(g, item) {
		return nil
	}
	player := item.Controller
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Eliminated {
		return nil
	}
	var cands []uuid.UUID
	for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			if (c.IsAura() || c.HasSubtype("Equipment")) && !c.IsToken() {
				cands = append(cands, c.InstanceID)
			}
		}
	}
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Destination: game.ChooseOntoBattlefield,
		Chooser:     player,
		Source:      danitha,
		Question:    "Danitha, Benalia's Hope — you may put an Aura or Equipment card from your hand or graveyard onto the battlefield attached to Danitha",
		Cards:       cands,
		Min:         0,
		Max:         1,
		ThenKey:     danithaPick,
		Carry:       []uuid.UUID{danitha},
	})
	return nil
}

// danithaPick puts the pick onto the battlefield attached to Danitha.
// Carry is Danitha. An on-disk key.
var danithaPick = game.RegisterChooseCardsThen("choose-cards/danitha-benalias-hope-attach",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Carry) == 0 || len(r.Picked) == 0 {
			return nil
		}
		if _, ok := g.LookupCardForEffect(r.Carry[0]); !ok || !onBattlefield(g, r.Carry[0]) {
			return nil
		}
		from := game.ZoneHand
		if z := g.FindCardZoneForEffect(r.Picked[0]); z != nil {
			from = z.Kind
		}
		return g.PutOntoBattlefieldTogetherKeyedForEffect(
			[]game.BatchEntry{{CardID: r.Picked[0], From: from}},
			game.ZoneEntryOptions{Controller: r.Chooser, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]}},
			game.KeyedTail{})
	})
