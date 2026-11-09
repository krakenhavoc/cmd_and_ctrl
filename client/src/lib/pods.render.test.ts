// @vitest-environment jsdom
//
// ADR 0220 slice 3: the pod pages (#/pods, #/pods/<id>, #/pods/join/<token>)
// and the words behind them.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Pod from "../routes/Pod.svelte";
import PodJoin from "../routes/PodJoin.svelte";
import Pods from "../routes/Pods.svelte";
import { L } from "./labels";
import { namesLine, pickableFriends, podLinkURL, suggestionLine } from "./pods";
import { setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const ME = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const BEA = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const CAL = "7a1d3c10-0f6e-4d46-8d1b-3a1f4d7a9c22";
const DEE = "8b2e4f60-1c3d-4e5f-9a7b-6c5d4e3f2a10";
const POD = "a1111111-1111-4111-8111-111111111111";

function sess(role: Session["principal"]["role"], userID?: string): Session {
  const expires = new Date(Date.now() + 86_400_000).toISOString();
  return {
    token: `${role}-tok`,
    expiresAt: expires,
    principal: { role, user_id: userID, name: "Ann", issued_at: expires, expires_at: expires },
  } as Session;
}

function reply(status: number, payload: unknown): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: "x",
    json: async () => payload,
    clone() {
      return this;
    },
  } as unknown as Response;
}

let calls: { method: string; path: string; body?: unknown }[] = [];
let role: "owner" | "member" = "owner";
let joinStatus = 200;
let suggestions: unknown[] = [];

function detail() {
  return {
    id: POD,
    name: "Thursday",
    role,
    owner: { user_id: ME, display_name: "Ann" },
    member_count: 2,
    max_members: 12,
    members: [
      { user_id: ME, display_name: "Ann", joined_at: 1789820000000, sharing: false, owner: true },
      { user_id: BEA, display_name: "Bea", joined_at: 1789830000000, sharing: true },
    ],
    invites: [
      {
        user_id: CAL,
        display_name: "Cal",
        invited_by: { display_name: "Ann" },
        since: 1,
        can_cancel: true,
      },
    ],
    ...(role === "owner" ? { link: { live: true, expires_at: 1790420000000 } } : {}),
  };
}

beforeEach(() => {
  calls = [];
  role = "owner";
  joinStatus = 200;
  suggestions = [];
  setSession(sess("identified", ME));
  vi.stubGlobal(
    "fetch",
    vi.fn(async (path: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      const body = init.body ? JSON.parse(init.body as string) : undefined;
      calls.push({ method, path, body });
      if (path === "/me/pods")
        return reply(200, {
          pods: [{ id: POD, name: "Thursday", member_count: 2, role: "owner" }],
          invites: [
            {
              pod_id: "b2",
              pod_name: "Friday",
              invited_by: { display_name: "Bea" },
              since: Date.now(),
            },
          ],
          suggestions: suggestions,
          max_pods: 10,
          max_members: 12,
        });
      if (path === `/pods/${POD}` && method === "GET") return reply(200, detail());
      if (path === "/me/friends")
        return reply(200, {
          friends: [
            { user_id: BEA, display_name: "Bea", since: 1 },
            { user_id: CAL, display_name: "Cal", since: 1 },
            { user_id: DEE, display_name: "Dee", since: 1 },
          ],
          incoming: [],
          outgoing: [],
        });
      if (path.startsWith("/pods/join/tok")) {
        if (joinStatus !== 200)
          return reply(joinStatus, { error: "This invite link can't be used." });
        if (method === "POST")
          return reply(200, { id: POD, name: "Thursday", member_count: 3, role: "member" });
        return reply(200, { name: "Thursday", owner_name: "Ann", member_count: 2, member: false });
      }
      if (path === "/pods" && method === "POST")
        return reply(201, { id: POD, name: body.name, member_count: 1, role: "owner" });
      if (path === `/pods/${POD}/link`)
        return method === "POST"
          ? reply(201, { token: "TOK", path: "#/pods/join/TOK", expires_at: 1790420000000 })
          : reply(204, null);
      if (method === "DELETE" || method === "PUT" || method === "PATCH" || method === "POST")
        return reply(204, null);
      return reply(404, { error: "unexpected " + path });
    }),
  );
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}
function button(root: HTMLElement, text: string): HTMLButtonElement | undefined {
  return [...root.querySelectorAll<HTMLButtonElement>("button")].find(
    (b) => b.textContent?.trim() === text,
  );
}

describe("pods words", () => {
  it("offers only friends not in the pod or invited", () => {
    const friends = [
      { user_id: BEA, display_name: "Bea" },
      { user_id: CAL, display_name: "Cal" },
      { user_id: DEE, display_name: "Dee" },
    ];
    const out = pickableFriends(friends, {
      members: [{ user_id: BEA, display_name: "Bea" }],
      invites: [{ user_id: CAL, display_name: "Cal", since: 1 }],
    });
    expect(out.map((f) => f.display_name)).toEqual(["Dee"]);
  });
  it("words a suggestion", () => {
    expect(namesLine(["Bea"])).toBe("Bea");
    expect(namesLine(["Bea", "Cal"])).toBe("Bea and Cal");
    expect(namesLine(["Bea", "Cal", "Dee"])).toBe("Bea, Cal and Dee");
    expect(suggestionLine({ games: 4, days: 30 })).toBe(
      "You’ve played 4 games together in the last 30 days.",
    );
  });
  it("builds the posted URL from the minted path", () => {
    expect(podLinkURL("#/pods/join/T", "https://x.test")).toBe("https://x.test/#/pods/join/T");
  });
});

describe("#/pods", () => {
  it("lists invites and pods, with an Owner tag", async () => {
    const { container } = render(Pods as never, {} as never);
    await settle();
    expect(container.querySelector("#p-invites")!.closest("section")!.textContent).toContain(
      "Bea invited you",
    );
    expect(container.querySelector(".pod-row")!.textContent).toContain("Thursday");
    expect(container.querySelector(".pod-row")!.textContent).toContain("Owner");
    expect(container.querySelector<HTMLAnchorElement>(".pod-link")!.getAttribute("href")).toBe(
      `#/pods/${POD}`,
    );
  });

  it("accepts an invite", async () => {
    const { container } = render(Pods as never, {} as never);
    await settle();
    click(button(container, L.acceptPodInvite)!);
    await settle();
    expect(calls.some((c) => c.method === "POST" && c.path === "/me/pod-invites/b2/accept")).toBe(
      true,
    );
    expect(container.textContent).toContain("You joined Friday.");
  });

  it("declines silently", async () => {
    const { container } = render(Pods as never, {} as never);
    await settle();
    click(button(container, L.declinePodInvite)!);
    await settle();
    expect(calls.some((c) => c.path === "/me/pod-invites/b2/decline")).toBe(true);
    expect(container.textContent).toContain("isn’t told");
  });

  it("New pod asks for a name, then makes it", async () => {
    const { container } = render(Pods as never, {} as never);
    await settle();
    click(button(container, L.newPod)!);
    flushSync();
    const input = container.querySelector<HTMLInputElement>("#pod-name")!;
    input.value = "  Friday  ";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    container.querySelector("form")!.dispatchEvent(new Event("submit", { cancelable: true }));
    await settle();
    expect(calls.find((c) => c.method === "POST" && c.path === "/pods")?.body).toEqual({
      name: "Friday",
    });
  });

  const SUGG = {
    key: "k".repeat(64),
    people: [{ display_name: "Bea" }, { display_name: "Cal" }],
    games: 4,
    last_at: 1,
    days: 30,
  };

  it("has no Suggested section without suggestions", async () => {
    const { container } = render(Pods as never, {} as never);
    await settle();
    expect(container.querySelector("#p-suggested")).toBeNull();
  });

  it("shows a suggestion, and Make this pod asks for a name then posts the key", async () => {
    suggestions = [SUGG];
    const { container } = render(Pods as never, {} as never);
    await settle();
    const card = container.querySelector("#p-suggested")!.closest("section")!;
    expect(card.textContent).toContain("Bea and Cal");
    expect(card.textContent).toContain("You’ve played 4 games together in the last 30 days.");
    click(button(container, L.makeSuggestedPod)!);
    flushSync();
    const input = container.querySelector<HTMLInputElement>(`#sg-name-${SUGG.key}`)!;
    input.value = "  Thursday  ";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    input.closest("form")!.dispatchEvent(new Event("submit", { cancelable: true }));
    await settle();
    expect(calls.find((c) => c.method === "POST" && c.path === "/pods")?.body).toEqual({
      name: "Thursday",
      suggestion: SUGG.key,
    });
  });

  it("Not now dismisses the suggestion by its key", async () => {
    suggestions = [SUGG];
    const { container } = render(Pods as never, {} as never);
    await settle();
    click(button(container, L.dismissPodSuggestion)!);
    await settle();
    expect(
      calls.some(
        (c) => c.method === "POST" && c.path === `/me/pod-suggestions/${SUGG.key}/dismiss`,
      ),
    ).toBe(true);
  });

  it("a guest is told to sign in", async () => {
    setSession(sess("player"));
    const { container } = render(Pods as never, {} as never);
    await settle();
    expect(container.textContent).toContain("Sign in with Discord");
    expect(calls.filter((c) => c.path.includes("pod"))).toHaveLength(0);
  });
});

describe("#/pods/<id>", () => {
  it("the owner sees members, tags, the link and manage controls", async () => {
    const { container } = render(Pod as never, { id: POD } as never);
    await settle();
    const t = container.textContent!;
    expect(t).toContain("Ann");
    expect(t).toContain("Sharing");
    expect(t).toContain("Not sharing");
    expect(t).toContain("Works until");
    expect(button(container, L.makePodLink)).toBeDefined();
    expect(button(container, "Rename")).toBeDefined();
    expect(button(container, L.leavePod)).toBeDefined();
  });

  it("a member has no link, rename or delete, and no menu on rows", async () => {
    role = "member";
    const { container } = render(Pod as never, { id: POD } as never);
    await settle();
    expect(button(container, L.makePodLink)).toBeUndefined();
    expect(button(container, "Rename")).toBeUndefined();
    expect(button(container, "Delete pod")).toBeUndefined();
    expect(container.querySelector(".more-btn")).toBeNull();
    expect(button(container, L.inviteToPod)).toBeDefined();
    expect(button(container, L.leavePod)).toBeDefined();
  });

  it("Remove asks first, offers a new link, and sends ?new_link=1", async () => {
    const { container } = render(Pod as never, { id: POD } as never);
    await settle();
    click(container.querySelector<HTMLButtonElement>(".more-btn")!);
    flushSync();
    const item = [...container.querySelectorAll<HTMLButtonElement>("[role=menuitem]")].find(
      (b) => b.textContent === L.removePodMember,
    )!;
    click(item);
    flushSync();
    expect(container.textContent).toContain("Remove Bea? They aren’t told.");
    expect(container.textContent).toContain("Also make a new invite link");
    expect(calls.some((c) => c.method === "DELETE")).toBe(false);
    const confirm = container.querySelector<HTMLElement>(".confirm")!;
    click(button(confirm, L.removePodMember)!);
    await settle();
    expect(
      calls.some(
        (c) => c.method === "DELETE" && c.path === `/pods/${POD}/members/${BEA}?new_link=1`,
      ),
    ).toBe(true);
  });

  it("Make owner hands the pod over", async () => {
    const { container } = render(Pod as never, { id: POD } as never);
    await settle();
    click(container.querySelector<HTMLButtonElement>(".more-btn")!);
    flushSync();
    click(
      [...container.querySelectorAll<HTMLButtonElement>("[role=menuitem]")].find(
        (b) => b.textContent === L.makePodOwner,
      )!,
    );
    await settle();
    expect(calls.find((c) => c.method === "PUT" && c.path === `/pods/${POD}/owner`)?.body).toEqual({
      user_id: BEA,
    });
  });

  it("the friend picker leaves out members and pending invitees", async () => {
    const { container } = render(Pod as never, { id: POD } as never);
    await settle();
    click(button(container, L.inviteToPod)!);
    await settle();
    const radios = [...container.querySelectorAll<HTMLInputElement>('input[name="friend"]')];
    expect(radios.map((r) => r.value)).toEqual([DEE]);
    radios[0].click();
    flushSync();
    click(button(container, "Send invite")!);
    await settle();
    expect(calls.find((c) => c.path === `/pods/${POD}/invites`)?.body).toEqual({ user_id: DEE });
  });

  it("a new link is shown once with Copy link", async () => {
    const { container } = render(Pod as never, { id: POD } as never);
    await settle();
    click(button(container, L.makePodLink)!);
    await settle();
    const box = container.querySelector<HTMLInputElement>('input[aria-label="Invite link"]')!;
    expect(box.value).toContain("#/pods/join/TOK");
    expect(button(container, L.copyPodLink)).toBeDefined();
  });

  it("Leave asks first, then leaves as me", async () => {
    const { container } = render(Pod as never, { id: POD } as never);
    await settle();
    click(button(container, L.leavePod)!);
    flushSync();
    expect(container.textContent).toContain("Leave Thursday?");
    click(button(container, L.leavePod)!);
    await settle();
    expect(calls.some((c) => c.method === "DELETE" && c.path === `/pods/${POD}/members/me`)).toBe(
      true,
    );
  });

  it("a pod that is not yours reads as gone", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => reply(404, { error: "no such pod" })),
    );
    const { container } = render(Pod as never, { id: POD } as never);
    await settle();
    expect(container.textContent).toContain("That pod isn’t there");
  });
});

describe("#/pods/join/<token>", () => {
  it("signed out shows only the sign-in line", async () => {
    setSession(null);
    const { container } = render(PodJoin as never, { token: "tok" } as never);
    await settle();
    expect(container.textContent).toContain("Sign in with Discord to join this pod");
    expect(container.textContent).not.toContain("Thursday");
    expect(calls.filter((c) => c.path.includes("pod"))).toHaveLength(0);
  });

  it("signed in previews, then Join is the yes", async () => {
    const { container } = render(PodJoin as never, { token: "tok" } as never);
    await settle();
    expect(container.textContent).toContain("Thursday");
    expect(container.textContent).toContain("Owned by Ann");
    expect(calls.some((c) => c.method === "POST")).toBe(false);
    click(button(container, L.joinPod)!);
    await settle();
    expect(calls.some((c) => c.method === "POST" && c.path === "/pods/join/tok")).toBe(true);
  });

  it("every dead link gets the one line", async () => {
    joinStatus = 404;
    const { container } = render(PodJoin as never, { token: "tok" } as never);
    await settle();
    expect(container.textContent).toContain("This invite link can’t be used.");
    expect(container.textContent).not.toContain("Thursday");
  });
});
