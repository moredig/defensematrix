import argparse
from textual.app import App, ComposeResult
from textual.widgets import Header, Footer, Log, Static
from textual.containers import Horizontal

from log_pipe import LogPipe

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
        self.connection_count = 0

    def compose(self) -> ComposeResult:
        yield Header(show_clock=True)

        with Horizontal(id="fleet-row"):
            yield Static("Fleet state is reported by the engine log.")

        yield Static(
            "Connections: 0   Hits: 0   Rotations: 0",
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
        if not self.log_file:
            return

        for event in self.pipe.read_from_file(self.log_file):
            message = event["message"]
            lowered = message.lower()
            if "hit" in lowered:
                self.hit_count += 1
            if "rotation" in lowered or "nuked" in lowered:
                self.rotation_count += 1
            if "connection tracked" in lowered:
                self.connection_count += 1
            self.write_log(message, event["level"])

        self.query_one("#stats-bar", Static).update(
            f"Connections: {self.connection_count}   Hits: {self.hit_count}   Rotations: {self.rotation_count}"
        )

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