package game

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// emblem.go is CR 114: the one object that exists only in the command
// zone, has no characteristics at all, and cannot be interacted with.
//
//	114.1  An emblem is a marker that represents an object that has no
//	       characteristics other than the abilities defined by the
//	       effect that created it.
//	114.2  Emblems exist in the command zone.
//	114.3  Abilities of emblems function in the command zone.
//	114.4  An emblem is neither a card nor a permanent.
//	114.5  "You get an emblem with [ability]" — the emblem is put into
//	       its owner's command zone, and its owner is its controller.
//
// ADR 0064 has the decisions. The short version:
//
//   - An emblem is an ordinary `Card` with a synthetic catalog key,
//     `"emblem:" + <the creating card's CatalogKey>`. That key IS its
//     identity, so CatalogLookup / CatalogStaticAbilities /
//     CatalogTriggers / CatalogAbilityKey work on it unchanged and no
//     field is added to Card.
//
//   - It lives in `Player.Emblems`, a SECOND command-zone slice beside
//     `Player.Command`. It is the command zone (CR 114.2); the split is
//     because every existing reader of Command.Cards means "commander
//     card" — the cast enumerator, the commander tax, the admin move
//     verb, ReplaceDeck, and the CR 704.5d token sweep that would
//     DELETE anything token-shaped sitting there.
//
//     The split is not a workaround, it is the enforcement: ZoneRef is
//     {Kind, Owner} with no discriminator, so zoneFromRefLocked,
//     routeDestinationLocked, castSourceZoneLocked, findCardZoneLocked
//     and findCardByIDLocked all resolve {command, owner} to
//     p.Command. An emblem therefore cannot be cast, moved, targeted,
//     sacrificed or destroyed, because no verb in the engine can name
//     it — which is exactly CR 114.
//
//   - Its statics join the layer pass's source list
//     (emblemContinuousEffectsLocked, called from
//     activeStaticAbilitiesLocked) and its triggers join the
//     harvester's zone walk (harvestFromEmblemsLocked, called from
//     triggerHarvester.OnEvent). One line each, no second pass.
//
//   - The only thing that removes one is CR 800.4a, its owner leaving
//     the game (leave_game.go).

// EmblemKeyPrefix marks a catalog key as an emblem's rather than a
// card's. A real Scryfall oracle ID is a UUID, so the prefix can never
// collide with one, and `effects.Register` files the emblem's CardDef
// under this key in the same map cards use.
const EmblemKeyPrefix = "emblem:"

// ErrNoEmblemRegistered is returned by CreateEmblemForEffect when the
// card whose ability is resolving declares no emblem. It means the
// card file is wrong — an ability said "you get an emblem" and no
// Spec.Emblem says what is on it — so it is an error rather than a
// silent no-op.
var ErrNoEmblemRegistered = errors.New("game: no emblem is registered for this source")

// EmblemKey is the catalog key of the emblem a card creates, derived
// from the card's own CatalogKey. A multi-face card gets the right one
// for free: a back face registers under "<oracle>#1", so its emblem is
// "emblem:<oracle>#1".
func EmblemKey(sourceKey string) string {
	if sourceKey == "" {
		return ""
	}
	return EmblemKeyPrefix + sourceKey
}

// IsEmblem reports whether this object is an emblem (CR 114) rather
// than a card or a token.
//
// It reads the catalog key rather than a flag on Card, because the key
// IS the identity: CR 114.1 says an emblem has no characteristics
// other than the abilities the effect defined, and the key is the
// handle on those abilities. A second source of truth would only be
// something for the first to disagree with later.
func (c Card) IsEmblem() bool {
	return strings.HasPrefix(c.OracleID, EmblemKeyPrefix)
}

// EmblemDef is the presentation half of an emblem's catalog entry: the
// name the board and the log give it, and the ability text a player
// reads on hover. The abilities themselves are the ordinary Static and
// Triggered slots of the same CardDef.
//
// Set only on an emblem's own def (the one filed under EmblemKey), so
// `CardDef.Emblem != nil` is how the engine tells an emblem's def from
// a card's.
type EmblemDef struct {
	// Label names the emblem: "Elspeth, Sun's Champion emblem".
	Label string
	// Text is the emblem's printed ability, verbatim.
	Text string
}

// EmblemView is one emblem as the projection and the tests read it.
// Derived, never stored: Label and Text come from the catalog on every
// read, so a text fix in a card file reaches a game already in
// progress.
type EmblemView struct {
	InstanceID uuid.UUID
	Label      string
	Text       string
}

// CreateEmblemForEffect is CR 114.5's "you get an emblem": the one
// seam that makes one. `sourceCardID` is the card whose ability is
// resolving; the emblem it creates is the one that card's Spec
// declares.
//
// The source is looked up in EVERY zone, not just the battlefield: a
// planeswalker can be killed in response to its own ultimate and the
// ability still resolves (CR 608.2), so the card is usually in a
// graveyard by the time this runs. CatalogKey is face-stable across
// that move for every single-faced card, and a card that changed face
// on the way out has already been reset by MoveCard (CR 712.8).
//
// Caller must hold g.mu in write mode — this is an effect-side helper
// and the resolution path already holds it.
func (g *Game) CreateEmblemForEffect(owner uuid.UUID, sourceCardID uuid.UUID) error {
	p := g.playerByIDLocked(owner)
	if p == nil {
		return ErrPlayerNotFound
	}
	src := g.findCardByIDLocked(sourceCardID)
	if src == nil {
		return ErrCardNotFound
	}
	key := EmblemKey(CatalogKey(*src))
	def := catalogDef(key)
	if def == nil || def.Emblem == nil {
		return ErrNoEmblemRegistered
	}
	if p.Emblems == nil {
		p.Emblems = newZone(ZoneCommand, p.ID)
	}
	// CR 114.1: every characteristic stays at its zero value — no
	// types, no mana cost, no colours, no P/T. `Name` is the board
	// label rather than a characteristic, and the empty TypeLine is
	// also what keeps the emblem clear of IsToken(), which is a
	// TypeLine test, and of the CR 704.5d sweep that runs it over the
	// command zone.
	p.Emblems.PushTop(Card{
		InstanceID: uuid.New(),
		Name:       def.Emblem.Label,
		OracleID:   key,
		Owner:      p.ID,
		Controller: p.ID,
		// CR 613.7c: a continuous effect from a static ability of an
		// emblem takes its timestamp when the emblem is created. This
		// is the object's CR 613.7 timestamp slot, under a
		// battlefield-flavoured name.
		EnteredBattlefieldAt: timeNowUnixNano(),
	})
	// Nothing emitted an event for the creation (ADR 0064 Decision 5),
	// so the invalidation is by hand — same shape as CR 800.4a's
	// removal in leave_game.go. An emblem with only triggers costs one
	// wasted recompute here and nothing after.
	g.layerVersion.Add(1)
	g.recomputeLayersLocked()
	return nil
}

// EmblemsForPlayer is the read side: every emblem a player has, with
// the label and text the catalog gives it. Returns nil for a player
// with none, which is nearly all of them.
//
// Caller must hold g.mu (read or write).
func (g *Game) EmblemsForPlayer(playerID uuid.UUID) []EmblemView {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Emblems == nil || len(p.Emblems.Cards) == 0 {
		return nil
	}
	out := make([]EmblemView, 0, len(p.Emblems.Cards))
	for _, c := range p.Emblems.Cards {
		v := EmblemView{InstanceID: c.InstanceID, Label: c.Name}
		if def := catalogDef(CatalogKey(c)); def != nil && def.Emblem != nil {
			v.Label = def.Emblem.Label
			v.Text = def.Emblem.Text
		}
		out = append(out, v)
	}
	return out
}

// emblemContinuousEffectsLocked is the emblem half of the layer pass's
// source gather (CR 114.3 — an emblem's abilities function in the
// command zone). It builds the same staticContinuousEffect adapter the
// battlefield and the turn-scoped registry use, so an emblem's static
// sorts into the same CR 613.7 timestamp order as everything else and
// needs no second application pass.
//
// `live` stays false, for the reason a turn-scoped static's does: only
// a permanent on the battlefield can be silenced by a CR 613.1f
// ability-removing effect, and nothing in the game can name an emblem
// to remove its abilities.
//
// The returned source pointers reference into p.Emblems.Cards with the
// same lifetime contract as the battlefield pointers: valid for one
// recompute pass, never retained across a mutation.
//
// Caller must hold g.mu in write mode (the recompute pass does).
func (g *Game) emblemContinuousEffectsLocked() []ContinuousEffect {
	if CatalogStaticAbilities == nil {
		return nil
	}
	var out []ContinuousEffect
	for _, p := range g.Seats {
		if p == nil || p.Emblems == nil {
			continue
		}
		for i := range p.Emblems.Cards {
			src := &p.Emblems.Cards[i]
			abilities := CatalogStaticAbilities(CatalogKey(*src))
			for _, ab := range abilities {
				out = append(out, staticContinuousEffect{
					ability:   ab,
					source:    src,
					timestamp: src.EnteredBattlefieldAt,
				})
			}
		}
	}
	return out
}

// harvestFromEmblemsLocked is the emblem half of the trigger
// harvester's walk. It is harvestFromZone over each seat's emblem
// zone, which is all an emblem's triggered ability needs: the harvest
// reads CatalogAbilityKey, runs AppliesTo against the source, and
// builds the item with the source's controller, so "whenever YOU draw
// a card" resolves ByYou against the emblem's owner with no special
// case anywhere.
//
// Caller must hold g.mu in write mode (the notify path does).
func (g *Game) harvestFromEmblemsLocked(pass *harvestPass) {
	for _, p := range g.Seats {
		if p == nil || p.Emblems == nil || len(p.Emblems.Cards) == 0 {
			continue
		}
		g.harvestFromZone(pass, p.Emblems)
	}
}
