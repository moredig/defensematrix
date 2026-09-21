import argparse
from textual.app import App, ComposeResult
from textual.widgets import Header, Footer, Log, Static
from textual.containers import Horizontal
from rich.text import Text

from renderer import render_boat_card, render_fleet_grid, render_header, render_stats
from log_pipe import LogPipe


# Mock fleet state — will be replaced by live Go engine output
MOCK_FLEET = [
    {"name": "boat-1", "role": "frontline",  "state": "IDLE",  "container_id": "a1b2c3d4e5f6"},
    {"name": "boat-2", "role": "shadow",     "state": "IDLE",  "container_id": "b2c3d4e5f6a1"},
    {"name": "boat-3", "role": "graveyard",  "state": "NUKED", "container_id": "c3d4e5f6a1b2"},
]


class BoatCard(Static):
    """A single boat status card widget."""

    def __init__(self, boat: dict, **kwargs):
        super().__init__(**kwargs)
        self.boat = boat

    def render(self):
        from renderer import STATE_COLORS
        color, label = STATE_COLORS.get(self.boat["state"], ("white", self.boat["state"]))
        return (
            f"[bold {color}]{self.boat['name'].upper()}[/bold {color}]\n"
            f"Role:  {self.boat['role'].upper()}\n"
            f"State: {label}\n"
            f"ID:    {self.boat['container_id'][:12]}"
        )


class DefenseMatrixDashboard(App):
    """Main Textual dashboard for the Defense Matrix."""

    CSS = """
    Screen {
        background: #0a0a0a;
    }

    #fleet-row {
        height: 10;
        margin: 1 2;
    }

    BoatCard {
        width: 1fr;
        height: 8;
        border: solid #1a1a2e;
        padding: 1 2;
        margin: 0 1;
        color: white;
    }

    #log-panel {
        height: 1fr;
        margin: 0 2;
        border: solid #1a1a2e;
        background: #050505;
    }

    #stats-bar {
        height: 3;
        margin: 0 2;
        background: #0d0d0d;
        border: solid #1a1a2e;
        padding: 0 2;
        color: cyan;
    }

    Header {
        background: #0d0d1a;
        color: ansi_bright_blue;
    }

    Footer {
        background: #0d0d1a;
    }
    """

    BINDINGS = [
        ("q", "quit", "Quit"),
        ("r", "refresh", "Refresh"),
    ]

    def __init__(self, log_file: str = None, **kwargs):
        super().__init__(**kwargs)
        self.log_file = log_file
        self.pipe = LogPipe(log_file)
        self.hit_count = 0
        self.rotation_count = 0
        self.fleet = MOCK_FLEET

    def compose(self) -> ComposeResult:
        yield Header(show_clock=True)

        with Horizontal(id="fleet-row"):
            for boat in self.fleet:
                yield BoatCard(boat, id=f"card-{boat['name']}")

        yield Static(
            f"⚡ Connections: 0   🎯 Hits: 0   🔄 Rotations: 0",
            id="stats-bar"
        )

        yield Log(id="log-panel", highlight=True)
        yield Footer()

    def on_mount(self) -> None:
        self.title = "DEFENSE MATRIX"
        self.sub_title = "Live Perimeter Monitor"
        self.set_interval(1, self.refresh_logs)
        self.write_log("[WAMAI] Defense Matrix dashboard online.", "BLUE")

    def refresh_logs(self) -> None:
        """Pulls new log events and updates the display."""
        if self.log_file:
            for event in self.pipe.read_from_file(self.log_file):
                self.write_log(event["message"], event["level"])
        else:
            import random
            mock_events = [
                ("[WAMAI] boat-1: connection from 45.33.32.156", "YELLOW"),
                ("[WAMAI] HIT — Pattern: SQL Injection | Level: HIGH", "RED"),
                ("[WAMAI] Rotation triggered — nuking frontline...", "RED"),
                ("[WAMAI] boat-2 promoted to frontline.", "BLUE"),
                ("[WAMAI] Recycler: rotation successful.", "BLUE"),
            ]
            event = random.choice(mock_events)
            self.write_log(event[0], event[1])

    def write_log(self, message: str, level: str = "INFO") -> None:
        """Writes a colored log line to the log panel."""
        log = self.query_one("#log-panel", Log)
        log.write_line(message)

    def action_refresh(self) -> None:
        self.refresh_logs()

    def action_quit(self) -> None:
        self.exit()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Defense Matrix Dashboard")
    parser.add_argument("--log", type=str, help="Path to Go engine log file", default=None)
    args = parser.parse_args()

    app = DefenseMatrixDashboard(log_file=args.log)
    app.run()