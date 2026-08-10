/**
 * Helm Pi Status Extension
 *
 * Writes Pi status to ~/.cache/helm-tmux/<session>.pi-status
 * so helm can display it in the session list.
 *
 * Events: session_start, agent_start, agent_end, session_shutdown
 *
 * Installation:
 *   1. Copy helm-tmux-pi-hook.sh to ~/.local/bin/
 *      cp hooks/helm-tmux-pi-hook.sh ~/.local/bin/
 *      chmod +x ~/.local/bin/helm-tmux-pi-hook.sh
 *
 *   2. Copy this file to ~/.pi/agent/extensions/
 *      cp hooks/helm-tmux-pi-status.ts ~/.pi/agent/extensions/
 *
 *   3. Restart Pi (or use /reload)
 */

import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";

export default function (pi: ExtensionAPI) {
  const hookScript = `${process.env.HOME}/.local/bin/helm-tmux-pi-hook.sh`;

  function callHook(event: string) {
    try {
      const { execSync } = require("child_process");
      execSync(`"${hookScript}" ${event}`, { stdio: "ignore" });
    } catch {
      // Non-fatal - hook script may not exist
    }
  }

  // Session starts - new agent context
  pi.on("session_start", async () => {
    callHook("start");
  });

  // Agent begins processing - working state
  pi.on("agent_start", async () => {
    callHook("working");
  });

  // Agent finishes (idle, waiting for user)
  pi.on("agent_end", async () => {
    callHook("waiting");
  });

  // Session ends - clean up status file
  pi.on("session_shutdown", async () => {
    callHook("end");
  });
}