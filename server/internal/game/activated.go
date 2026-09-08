package game

import (
	"github.com/google/uuid"
)

// activated.go — S21 sub-PR 2: activated abilities in the catalog
// (CR 602), the fourth ability type after triggered (S19), static
// (S16) and replacement (S17).
//
// Before this, a catalog card's activated ability didn't exist:
// ActivateAbility put a labelled item on the stack with no cost
// check and no effect, and players resolved it in their heads. A
// sacrifice outlet ("Sacrifice a creature: deal 1 damage") can't
// work that way — the cost IS the game action.
//
// The shape mirrors the rest of the catalog surface. A card declares
// its abilities; the engine validates timing and cost, pays the
// cost, and pushes a stack item carrying the same Effect closure
// S19's triggered abilities use. Resolution is therefore already
// written: resolveTopAbilityLocked runs Effect with the live game,
// after the CR 608.2b target re-check.
//
// Mana abilities stay separate (CR 605.3a — they don't use the
// stack) and keep their own ManaAbilityShape path. An ability that
// would be a mana ability but whose cost sacrifices a DIFFERENT
// permanent (Ashnod's Altar) fits neither surface cleanly and is
// deferred; the note is in ADR 0020.

// AbilityCost is what a player pays to activate an ability
// (CR 602.1b). Every field is additive: Goblin Bombardment is
// {SacrificeOther: creature-spec}, Krenko is {Tap: true}, a
// hypothetical "{2}, {T}, Sacrifice a creature:" would set all
// three.
type AbilityCost struct {
	// Tap requires the source to be untapped, taps it, and — for a
	// creature source — enforces summoning sickness (CR 302.1).
	Tap bool

	// SacrificeSelf sacrifices the source as part of the cost.
	SacrificeSelf bool

	// SacrificeOther sacrifices one other permanent the activator
	// controls, chosen at announce and matched against this spec
	// ("Sacrifice a creature"). The activator names it in
	// ActivateAbilityParams.SacrificeIDs. The source itself is a
	// legal choice when the spec admits it — Carrion Feeder can eat
	// itself, as in paper.
	SacrificeOther *TargetSpec

	// Mana is a printed cost string ("{1}{B}") paid from the pool
	// under the same strict / permissive rules casting uses. Empty
	// means no mana component.
	Mana string

	// Life is a life payment (CR 118.8). Paying life is legal at any
	// total above the payment; the SBA loop handles the rest.
	Life int
}

// ActivatedAbilityShape is one activated ability on a permanent, as
// the game package consumes it. Mirrors effects.ActivatedAbility,
// which lives in the catalog package (the same import-cycle dodge
// ManaAbilityShape uses).
type ActivatedAbilityShape struct {
	// Label is the oracle text of the ability, shown in the
	// activation menu: "Sacrifice a creature: Goblin Bombardment
	// deals 1 damage to any target."
	Label string

	Cost AbilityCost

	// Targets is the ability's target clause, validated at announce
	// and re-checked at resolution exactly as a spell's is. Nil for
	// untargeted abilities.
	Targets *TargetSpec

	// SorcerySpeed marks "activate only as a sorcery" (CR 602.5d).
	SorcerySpeed bool

	// Effect runs at resolution against the live game. Same contract
	// as TriggeredAbility's stack items: never capture a *Card,
	// read what you need off the item and the game.
	Effect func(g *Game, item *StackItem) error
}

// CatalogActivatedAbilities returns the registered activated
// abilities for an oracle ID, or nil. Populated by the effects
// package at init, alongside the other catalog hooks.
var CatalogActivatedAbilities func(oracleID string) []ActivatedAbilityShape

// ActivatedAbilitiesForCard returns the activated abilities a card
// offers right now. Catalog-only: unlike mana abilities there's no
// synthetic fallback, because there's no ability every card of some
// type implicitly has.
func ActivatedAbilitiesForCard(c Card) []ActivatedAbilityShape {
	if CatalogActivatedAbilities == nil || c.OracleID == "" {
		return nil
	}
	return CatalogActivatedAbilities(c.OracleID)
}

// ActivateAbilityParams carries the announce-time choices for a
// catalog activated ability.
type ActivateAbilityParams struct {
	// SacrificeIDs names the permanents paid to a SacrificeOther
	// cost, in the order the activator picked them. Exactly one
	// today (no catalog card sacrifices two), but a slice so the
	// wire shape survives contact with Altar of Dementia-style
	// costs.
	SacrificeIDs []uuid.UUID

	// Targets are the ability's targets, validated against the
	// ability's spec.
	Targets []TargetRef

	// Strict / AutoTap mirror CastSpellParams: they gate the mana
	// component of the cost the same way a cast is gated.
	Strict  bool
	AutoTap bool
}

// ActivateCatalogAbility activates ability `index` on a permanent
// the player controls: validates timing and cost, pays the cost,
// and pushes a stack item carrying the ability's effect (CR 602.2).
// The activator retains priority, as with any announce.
//
// Costs are validated in full before ANY of them is paid, so a
// half-paid activation can't strand the board — the same discipline
// the sacrifice-cost mana ability path uses.
//
// Caller must NOT hold g.mu.
func (g *Game) ActivateCatalogAbility(playerID, cardID uuid.UUID, index int, params ActivateAbilityParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.SplitSecondActive {
		return ErrSplitSecondActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	source := findBattlefieldCard(g, cardID)
	if source == nil {
		return ErrCardNotFound
	}
	if source.Controller != playerID {
		return ErrCardCallerMismatch
	}
	abilities := ActivatedAbilitiesForCard(*source)
	if index < 0 || index >= len(abilities) {
		return ErrInvalidParam
	}
	ab := abilities[index]

	// --- timing -------------------------------------------------
	if ab.SorcerySpeed && !g.sorcerySpeedOpenLocked(playerID) {
		return ErrSorcerySpeedRequired
	}

	// --- validate every cost before paying any ------------------
	if ab.Cost.Tap {
		if source.Tapped {
			return ErrAlreadyTapped
		}
		// CR 302.1: a creature's {T} ability needs it to have been
		// under your control since your most recent turn began.
		// Non-creature sources (Krenko is a creature; an artifact
		// with {T} isn't) are never sick.
		if source.IsCreature() && HasSummoningSickness(source) {
			return ErrSummoningSick
		}
	}
	sacrifices, err := g.validateSacrificeCostLocked(playerID, cardID, ab.Cost, params.SacrificeIDs)
	if err != nil {
		return err
	}
	if ab.Cost.Life > 0 && p.Life < ab.Cost.Life {
		// CR 118.8 forbids paying more life than you have. Paying
		// down to exactly 0 is legal; the SBA loop ends the game
		// after.
		return ErrInvalidParam
	}
	if ab.Targets != nil {
		if err := g.validateTargetsLocked(playerID, ab.Targets, params.Targets); err != nil {
			return err
		}
	} else if len(params.Targets) > 0 {
		return ErrInvalidParam
	}

	// --- pay ----------------------------------------------------
	if ab.Cost.Mana != "" {
		if err := g.payAbilityManaCostLocked(p, cardID, ab.Cost.Mana, params); err != nil {
			return err
		}
	}
	if ab.Cost.Tap {
		source.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: cardID})
	}
	if ab.Cost.Life > 0 {
		if err := g.ChangePlayerLifeForEffect(cardID, playerID, -ab.Cost.Life); err != nil {
			return err
		}
	}
	// Sacrifices last: they move cards, which invalidates `source`.
	for _, id := range sacrifices {
		if err := g.sacrificePermanentLocked(id); err != nil {
			return err
		}
	}
	source = nil

	// --- announce -----------------------------------------------
	itemID := uuid.New()
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	item := &StackItem{
		ID:           itemID,
		Kind:         StackItemActivated,
		Controller:   playerID,
		Owner:        playerID,
		SourceCardID: cardID,
		Label:        ab.Label,
		Targets:      append([]TargetRef(nil), params.Targets...),
		Effect:       ab.Effect,
		targetSpec:   ab.Targets,
		Seq:          g.nextStackSeqLocked(),
	}
	g.StackMeta[itemID] = item
	g.EmitEvent(Event{
		Kind:   EventTrigger,
		Actor:  playerID,
		Source: cardID,
		CardID: cardID,
	})
	// The cost may have queued dies-triggers (a sacrifice outlet
	// feeding Blood Artist). Drain them so they sit ABOVE the
	// ability on the stack, which is where paying a cost puts them.
	g.runStateChecksLocked()
	return nil
}

// validateSacrificeCostLocked resolves the SacrificeSelf /
// SacrificeOther components into the concrete list of permanents to
// sacrifice, without moving anything. Caller must hold g.mu.
func (g *Game) validateSacrificeCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost, chosen []uuid.UUID) ([]uuid.UUID, error) {
	var out []uuid.UUID
	if cost.SacrificeSelf {
		out = append(out, sourceID)
	}
	if cost.SacrificeOther == nil {
		if len(chosen) > 0 {
			return nil, ErrInvalidParam
		}
		return out, nil
	}
	if len(chosen) != 1 {
		return nil, ErrInvalidParam
	}
	id := chosen[0]
	c := findBattlefieldCard(g, id)
	if c == nil {
		return nil, ErrCardNotFound
	}
	// CR 701.17b — you can only sacrifice what you control.
	if c.Controller != playerID {
		return nil, ErrCardCallerMismatch
	}
	if !g.targetLegalLocked(playerID, cost.SacrificeOther, TargetRef{Kind: TargetCard, ID: id}) {
		return nil, ErrIllegalTarget
	}
	// Paying the same permanent twice (self-sacrifice plus the same
	// card as the "other") isn't a legal cost payment.
	for _, already := range out {
		if already == id {
			return nil, ErrInvalidParam
		}
	}
	return append(out, id), nil
}

// payAbilityManaCostLocked charges an activated ability's mana
// component through the same pool / auto-tap machinery a cast uses.
// Permissive mode (the default) leaves the pool alone and emits a
// cost warning, matching how S15 treats an unaffordable cast.
// Caller must hold g.mu.
func (g *Game) payAbilityManaCostLocked(p *Player, sourceID uuid.UUID, costStr string, params ActivateAbilityParams) error {
	cost, err := ParseCost(costStr)
	if err != nil {
		return ErrInvalidParam
	}
	if !params.Strict && !params.AutoTap {
		if !p.ManaPool.CanPay(cost, 0) {
			g.EmitEvent(Event{
				Kind:   EventCostWarning,
				Actor:  p.ID,
				Source: sourceID,
			})
			return nil
		}
		p.ManaPool.SpendMana(cost, 0)
		return nil
	}
	if params.AutoTap && !p.ManaPool.CanPay(cost, 0) {
		plan, ok := g.autoTapLocked(p.ID, cost, 0, nil)
		if !ok {
			return &InsufficientManaError{Missing: p.ManaPool.Missing(cost, 0)}
		}
		g.materializePlanLocked(p, plan, cost)
	}
	if !p.ManaPool.CanPay(cost, 0) {
		return &InsufficientManaError{Missing: p.ManaPool.Missing(cost, 0)}
	}
	p.ManaPool.SpendMana(cost, 0)
	return nil
}
