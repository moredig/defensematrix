import argparse
import os
import sys
import time

from log_pipe import LogPipe


def clear_screen():
    os.system("cls" if os.name == "nt" else "clear")


def render(events, counters, width=100):
    lines = []
    lines.append("DEFENSE MATRIX")
    lines.append("=" * width)
    lines.append(f"Frontline: {counters.get('frontline', 'unknown')} | Shadow: {counters.get('shadow', 'unknown')} | Graveyard: {counters.get('graveyard', 'unknown')}")
    lines.append(f"Connections: {counters.get('connections', 0)} | Hits: {counters.get('hits', 0)} | Rotations: {counters.get('rotations', 0)}")
    lines.append("-" * width)
    lines.append("Recent events:")

    for event in events[-25:]:
        msg = event.get("message", "")
        lines.append(f"{event.get('timestamp', '')} {msg}")

    if not events:
        lines.append("Waiting for engine output...")

    sys.stdout.write("\033[2J\033[H")
    sys.stdout.write("\n".join(lines) + "\n")
    sys.stdout.flush()


def parse_counters(message: str, counters: dict) -> None:
    lower = message.lower()
    if "connection tracked" in lower:
        counters["connections"] = counters.get("connections", 0) + 1
    if "hit" in lower:
        counters["hits"] = counters.get("hits", 0) + 1
    if "rotation" in lower or "nuked" in lower:
        counters["rotations"] = counters.get("rotations", 0) + 1
    if "promoted to frontline" in lower:
        counters["frontline"] = "boat-2"
    if "sent to graveyard" in lower:
        counters["graveyard"] = "boat-3"


def main():
    parser = argparse.ArgumentParser(description="Defense Matrix terminal monitor")
    parser.add_argument("--log", type=str, default=None, help="Path to the engine log file")
    parser.add_argument("--refresh", type=float, default=1.0, help="Refresh interval in seconds")
    parser.add_argument("--once", action="store_true", help="Render one snapshot and exit")
    args = parser.parse_args()

    pipe = LogPipe(args.log)
    events = []
    counters = {
        "frontline": "boat-1",
        "shadow": "boat-2",
        "graveyard": "boat-3",
        "connections": 0,
        "hits": 0,
        "rotations": 0,
    }

    try:
        while True:
            for event in pipe.read_from_file(args.log) if args.log else pipe.read_from_stdin():
                events.append(event)
                parse_counters(event.get("message", ""), counters)

            render(events, counters)
            if args.once:
                break
            time.sleep(args.refresh)
    except KeyboardInterrupt:
        print("\nStopped.")


if __name__ == "__main__":
    main()