package bot

// access.go is ADR 0156's 2026-10-16 amendment (#3325): access requests
// with Approve and Decline buttons in the community server's
// #account-requests channel.
//
// Discord's onboarding puts a member who answers "request account access"
// in that channel, and the site's copy tells them to say hi there. The
// first message a member without the player role posts there gets a reply
// from the bot: an access request with two buttons. A mod presses one:
//
//   - Approve grants the player role (Elder Dragon Highlander on the C2
//     server, CMDCTRL_DISCORD_PLAYER_ROLE_ID), the role the server's
//     sign-up gate checks, and the request says who approved it.
//   - Decline changes nothing on Discord, and the request says who
//     declined it.
//
// Either way the buttons go, and a line goes to #mod-log when it is set.
//
// "A mod" is someone who could grant the role by hand: a member with
// Manage Roles (or Administrator) in that channel, or someone on the
// bot's admin lists (CMDCTRL_DISCORD_ADMIN_USER_IDS / _ROLE_IDS). The
// button does nothing that person could not already do.
//
// The buttons are stateless: the custom ID carries the action and the
// member's id, so a request posted before a bot restart still works.
// What is in memory is only which members already have a request
// posted, so a member who keeps chatting gets one request a day, not one
// per message. A restart forgets it; at worst a member gets a second
// request.
//
// Needs, on the Discord side (owner setup in the ADR): the bot's role
// has Manage Roles and sits ABOVE the player role, and the bot can read
// and post in #account-requests.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	// accessCustomIDPrefix namespaces the two buttons' custom IDs:
	// "c2-access:approve:<user id>" and "c2-access:decline:<user id>".
	accessCustomIDPrefix = "c2-access:"
	accessApprove        = "approve"
	accessDecline        = "decline"

	// accessRequestCooldown is how long after a request is posted a
	// member's further messages in the channel post nothing more.
	accessRequestCooldown = 24 * time.Hour
	// accessRequestMaxTracked caps the members remembered; past it the
	// expired entries are dropped, and if none are, the oldest.
	accessRequestMaxTracked = 2000

	// accessRequestColor is the request embed's side bar: blurple while
	// it waits, green once approved, grey once declined.
	accessRequestColor  = 0x5865F2
	accessApprovedColor = 0x57F287
	accessDeclinedColor = 0x99AAB5
	accessModLogColor   = 0xFEE75C
)

// accessRequests remembers when each member's last request was posted.
type accessRequests struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

func newAccessRequests() *accessRequests {
	return &accessRequests{seen: map[string]time.Time{}}
}

// claim reports whether a request may be posted for userID now, and if
// so records it.
func (a *accessRequests) claim(userID string, now time.Time) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if at, ok := a.seen[userID]; ok && now.Sub(at) < accessRequestCooldown {
		return false
	}
	if len(a.seen) >= accessRequestMaxTracked {
		a.pruneLocked(now)
	}
	a.seen[userID] = now
	return true
}

// release forgets userID, so a request that could not be posted is
// tried again on the member's next message.
func (a *accessRequests) release(userID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.seen, userID)
}

func (a *accessRequests) pruneLocked(now time.Time) {
	var oldestID string
	var oldest time.Time
	for id, at := range a.seen {
		if now.Sub(at) >= accessRequestCooldown {
			delete(a.seen, id)
			continue
		}
		if oldestID == "" || at.Before(oldest) {
			oldestID, oldest = id, at
		}
	}
	if len(a.seen) >= accessRequestMaxTracked && oldestID != "" {
		delete(a.seen, oldestID)
	}
}

// HandleAccessRequestMessage is the MessageCreate callback for
// #account-requests. It reads who posted and their roles, never the
// message's text, so it needs no Message Content intent.
func (h *Handler) HandleAccessRequestMessage(s messageSender, m *discordgo.MessageCreate) {
	if !h.cfg.AccessRequestsEnabled() || m == nil || m.Message == nil || m.Author == nil {
		return
	}
	if m.Author.Bot || m.Author.System || m.WebhookID != "" {
		return
	}
	// A join notice, a pin or a thread start is not someone asking.
	if m.Type != discordgo.MessageTypeDefault && m.Type != discordgo.MessageTypeReply {
		return
	}
	if m.ChannelID != h.cfg.AccessRequestChannelID || !h.cfg.GuildAllowed(m.GuildID) {
		return
	}
	var roles []string
	if m.Member != nil {
		roles = m.Member.Roles
	}
	// Already admitted, or a mod chatting in the channel.
	if hasRole(roles, h.cfg.PlayerRoleID) || h.cfg.isAdminID(m.Author.ID, roles) {
		return
	}
	if !h.access.claim(m.Author.ID, h.now()) {
		return
	}
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, accessRequestMessage(m.Author.ID, m.Reference())); err != nil {
		h.access.release(m.Author.ID)
		h.log.Warn("access request not posted", "error", err.Error())
		return
	}
	h.log.Info("access request posted", "member", m.Author.ID)
}

// accessRequestMessage is the request: a reply to the member's message
// that mentions them without pinging anyone, and the two buttons.
func accessRequestMessage(userID string, ref *discordgo.MessageReference) *discordgo.MessageSend {
	return &discordgo.MessageSend{
		Reference:       ref,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
		Embeds: []*discordgo.MessageEmbed{{
			Title: "Access request",
			Description: fmt.Sprintf("<@%s> is asking for a CMD & CTRL account. "+
				"Approve gives them the role that lets them sign in. Only a mod can answer.", userID),
			Color: accessRequestColor,
		}},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "Approve", Style: discordgo.SuccessButton, CustomID: accessCustomID(accessApprove, userID)},
				discordgo.Button{Label: "Decline", Style: discordgo.DangerButton, CustomID: accessCustomID(accessDecline, userID)},
			}},
		},
	}
}

func accessCustomID(action, userID string) string {
	return accessCustomIDPrefix + action + ":" + userID
}

// parseAccessCustomID reads a button's action and member back.
func parseAccessCustomID(customID string) (action, userID string, ok bool) {
	rest, found := strings.CutPrefix(customID, accessCustomIDPrefix)
	if !found {
		return "", "", false
	}
	action, userID, found = strings.Cut(rest, ":")
	if !found || (action != accessApprove && action != accessDecline) || !isSnowflake(userID) {
		return "", "", false
	}
	return action, userID, true
}

// mayModerate reports whether the person who pressed a button may answer
// an access request: a member who could grant the role by hand (Manage
// Roles or Administrator, as Discord computed it for this channel), or
// someone on the bot's admin lists.
func (h *Handler) mayModerate(i *discordgo.Interaction) bool {
	if h.cfg.IsAdmin(i) {
		return true
	}
	if i == nil || i.Member == nil {
		return false
	}
	return i.Member.Permissions&(discordgo.PermissionManageRoles|discordgo.PermissionAdministrator) != 0
}

// roleGranter is the Discord call Approve makes. *discordgo.Session
// satisfies it.
type roleGranter interface {
	GuildMemberRoleAdd(guildID, userID, roleID string, options ...discordgo.RequestOption) error
}

// interactionResponder answers a button press. *discordgo.Session
// satisfies it.
type interactionResponder interface {
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
}

// accessSession is everything an Approve or Decline press touches.
type accessSession interface {
	roleGranter
	interactionResponder
	messageSender
}

// dispatchAccessComponent handles an Approve or Decline press.
func (h *Handler) dispatchAccessComponent(ctx context.Context, s accessSession, i *discordgo.InteractionCreate, customID string) {
	action, userID, ok := parseAccessCustomID(customID)
	if !ok {
		h.log.Warn("malformed access request custom id", "custom_id", customID)
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse("This button is out of date."))
		return
	}
	if !h.cfg.AccessRequestsEnabled() {
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(
			"Access requests are turned off on this bot ("+AccessRequestChannelEnv+" and "+PlayerRoleEnv+").",
		))
		return
	}
	if !h.mayModerate(i.Interaction) {
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse("Only a mod can approve or decline an access request."))
		return
	}
	modID, _ := interactionInvoker(i.Interaction)
	modName := memberDisplayName(i.Interaction)

	if action == accessDecline {
		_ = s.InteractionRespond(i.Interaction, accessAnsweredResponse(userID,
			fmt.Sprintf("Declined by <@%s>.", modID), accessDeclinedColor))
		h.log.Info("access request declined", "member", userID, "mod", modID)
		h.postAccessModLog(s, "Access request declined", userID, modID)
		return
	}

	err := s.GuildMemberRoleAdd(i.GuildID, userID, h.cfg.PlayerRoleID,
		discordgo.WithContext(ctx),
		discordgo.WithAuditLogReason("CMD & CTRL access request approved by "+modName))
	switch {
	case err == nil:
		_ = s.InteractionRespond(i.Interaction, accessAnsweredResponse(userID,
			fmt.Sprintf("Approved by <@%s>. <@%s>, you can sign in to CMD & CTRL now.", modID, userID), accessApprovedColor))
		h.log.Info("access request approved", "member", userID, "mod", modID)
		h.postAccessModLog(s, "Access request approved", userID, modID)
	case restErrorCode(err) == discordgo.ErrCodeUnknownMember:
		_ = s.InteractionRespond(i.Interaction, accessAnsweredResponse(userID,
			"They left the server before a mod answered.", accessDeclinedColor))
	case restStatus(err) == http.StatusForbidden:
		h.log.Error("access request: the bot cannot grant the player role", "error", err.Error())
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(approveForbiddenMessage))
	default:
		h.log.Error("access request: granting the player role failed", "error", err.Error())
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse("Discord didn't grant the role. Try again in a moment."))
	}
}

// approveForbiddenMessage is what a mod sees when Discord refuses the
// grant: the owner setup the ADR lists, in one line.
const approveForbiddenMessage = "The bot isn't allowed to give that role. " +
	"In Server Settings → Roles, give the bot's role Manage Roles and drag it above the player role, then press Approve again."

// accessAnsweredResponse edits the request in place: the outcome under
// the original line, and no buttons.
func accessAnsweredResponse(userID, outcome string, color int) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			AllowedMentions: &discordgo.MessageAllowedMentions{},
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "Access request",
				Description: fmt.Sprintf("<@%s> asked for a CMD & CTRL account.\n%s", userID, outcome),
				Color:       color,
			}},
			Components: []discordgo.MessageComponent{},
		},
	}
}

// postAccessModLog adds the outcome to #mod-log, when it is set. It
// mentions the member and the mod, which never notifies anyone in an
// embed, and carries nothing else.
func (h *Handler) postAccessModLog(s messageSender, title, userID, modID string) {
	if h.cfg.ModLogChannelID == "" {
		return
	}
	msg := &discordgo.MessageSend{
		AllowedMentions: &discordgo.MessageAllowedMentions{},
		Embeds: []*discordgo.MessageEmbed{{
			Title: title,
			Color: accessModLogColor,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Member", Value: "<@" + userID + ">", Inline: true},
				{Name: "By", Value: "<@" + modID + ">", Inline: true},
			},
			Timestamp: h.now().UTC().Format(time.RFC3339),
		}},
	}
	if _, err := s.ChannelMessageSendComplex(h.cfg.ModLogChannelID, msg); err != nil {
		h.log.Warn("mod log entry not posted", "error", err.Error())
	}
}

// memberDisplayName is the presser's name for the audit log: server
// nickname, then global name, then username.
func memberDisplayName(i *discordgo.Interaction) string {
	if i == nil {
		return "a mod"
	}
	if i.Member != nil {
		if i.Member.Nick != "" {
			return i.Member.Nick
		}
		if i.Member.User != nil {
			if i.Member.User.GlobalName != "" {
				return i.Member.User.GlobalName
			}
			if i.Member.User.Username != "" {
				return i.Member.User.Username
			}
		}
	}
	return "a mod"
}

// hasRole reports whether roles holds want.
func hasRole(roles []string, want string) bool {
	for _, r := range roles {
		if r == want {
			return true
		}
	}
	return false
}

// restStatus is the HTTP status of a Discord REST error, 0 for any
// other error.
func restStatus(err error) int {
	var re *discordgo.RESTError
	if errors.As(err, &re) && re.Response != nil {
		return re.Response.StatusCode
	}
	return 0
}

// restErrorCode is Discord's JSON error code of a REST error, 0 when
// there is none.
func restErrorCode(err error) int {
	var re *discordgo.RESTError
	if errors.As(err, &re) && re.Message != nil {
		return re.Message.Code
	}
	return 0
}
