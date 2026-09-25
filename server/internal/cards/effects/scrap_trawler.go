package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrap Trawler — Artifact Creature — Construct {3}, 2/2 (#300, #1223):
//
//	"Whenever this creature dies or another artifact you control is
//	 put into a graveyard from the battlefield, return to your hand
//	 target artifact card in your graveyard with lesser mana value."
//
// The card the trigger-data seam was filed on, and the shortest
// statement of it: "LESSER" is lesser than the mana value of the
// artifact that just died, and `TriggeredAbility.Targets` is static
// catalog data declared before the game started. There was nowhere
// for the number to come from.
//
// `TargetsFrom` is where it comes from now: the clause is built when
// the trigger is put on the stack (CR 603.3d), from the same
// CR 603.10 last-known-information snapshot the rules say an
// LTB-triggered ability is judged on. A 4-mana artifact dying offers
// the 3-drops and the 2-drops; a Scrap Trawler chain resolves in
// descending mana value exactly as it does in paper, because each
// link's clause is cut to its own link.
//
// CR 603.3d does the rest: when nothing in the graveyard is cheaper,
// the ability is not put on the stack at all and nobody is prompted.
func init() {
	Register(Spec{
		OracleID:     "164f3f85-21fc-40b7-9871-4f303ba98428",
		Name:         "Scrap Trawler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventLTB},
			AppliesTo: scrapTrawlerArtifactHitTheYard,
			// TargetsFrom, not Targets: the clause is a fact about
			// what just died. See trigger_event.go.
			TargetsFrom: scrapTrawlerCheaperArtifactClause,
			Key:         "Scrap Trawler — return a cheaper artifact card to your hand",
			Effect:      b15ReturnListedCardsFromGraveyardToHand,
		}},
	})
}

// scrapTrawlerArtifactHitTheYard is the trigger condition: the source
// itself died, or another artifact its controller controlled was put
// into a graveyard from the battlefield.
//
// "Put into a graveyard from the battlefield" rather than "dies" for
// the non-creature half — they are the same event here (CR 700.4
// defines "dies" for creatures only) and EventLTB with a graveyard
// destination is how the engine says both.
//
// The controller test reads the card where it now is, which is the
// graveyard: control is not a characteristic that survives the move
// in the rules, but Card.Controller is not cleared by it and is the
// only reading available to an AppliesTo. It is the same reading
// every dies-trigger in the catalog makes (helpers.go's diedCreature)
// and it is right for every board that does not change control of a
// permanent in the same event that destroys it.
func scrapTrawlerArtifactHitTheYard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsArtifact() && c.Controller == source.Controller
}

// scrapTrawlerCheaperArtifactClause builds the target clause from the
// triggering event: "target artifact card in your graveyard with
// lesser mana value".
//
// The ceiling is the mana value of the object that LEFT, read off the
// trigger's CR 603.10 snapshot rather than off the card now sitting
// in the graveyard. For most boards the two agree; they part company
// for exactly the objects the rule was written about — a copied
// artifact, one that was animated and then destroyed, a face-down
// permanent — and the snapshot is the one the rules ask for.
//
// A zero ceiling returns a clause nothing can satisfy rather than
// nil: "lesser than 0" has no answers, and CR 603.3d removes the
// ability. Returning nil would instead say "this trigger targets
// nothing", which would put an ability on the stack that returns
// nothing to nobody.
func scrapTrawlerCheaperArtifactClause(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	ceiling := 0
	if tc.Object != nil {
		ceiling = tc.Object.ManaValue
	}
	return TargetCardInGraveyard(
		"target artifact card in your graveyard with lesser mana value",
		YouOwn(), Artifact(), ManaValueLE(ceiling-1),
	)
}
