from rich.console import Console
from rich.table import Table
from rich.text import Text
from rich.panel import Panel
from rich.columns import Columns
from datetime import datetime


# State color map
STATE_COLORS = {
    "IDLE":        ("blue",   "🔵 IDLE"),
    "PROBING":     ("yellow", "🟡 PROBING"),
    "INFILTRATED": ("orange1","🟠 INFILTRATED"),
    "NUKED":       ("red",    "🔴 NUKED"),
}

LOG_COLORS = {
    "RED":    "bold red",
    "ORANGE": "bold orange1",
    "YELLOW": "bold yellow",
    "BLUE":   "bold blue",
    "INFO":   "dim white",
}


console = Console()


def render_boat_card(name: str, role: str, state: str, container_id: str = "N/A") -> Panel:
    """Renders a single boat status card."""
    color, label = STATE_COLORS.get(state, ("white", state))

    content = Text()
    content.append(f"  Role:      ", style="bold white")
    content.append(f"{role.upper()}\n", style="bold cyan")
    content.append(f"  State:     ", style="bold white")
    content.append(f"{label}\n", style=f"bold {color}")
    content.append(f"  Container: ", style="bold white")
    content.append(f"{container_id[:12] if container_id != 'N/A' else 'N/A'}\n", style="dim white")
    content.append(f"  Updated:   ", style="bold white")
    content.append(f"{datetime.now().strftime('%H:%M:%S')}", style="dim white")

    return Panel(
        content,
        title=f"[bold {color}]{name.upper()}[/bold {color}]",
        border_style=color,
        expand=True,
    )


def render_fleet_grid(boats: list) -> None:
    """Renders all three boat cards side by side."""
    cards = [
        render_boat_card(
            name=b["name"],
            role=b["role"],
            state=b["state"],
            container_id=b.get("container_id", "N/A"),
        )
        for b in boats
    ]
    console.print(Columns(cards, equal=True, expand=True))


def render_log_line(event: dict) -> None:
    """Prints a single log event with color coding."""
    color = LOG_COLORS.get(event["level"], "white")
    console.print(
        f"[dim]{event['timestamp']}[/dim] [{color}]{event['message']}[/{color}]"
    )


def render_header() -> None:
    """Prints the Defense Matrix header banner."""
    console.print(Panel(
        Text("DEFENSE MATRIX — LIVE PERIMETER", style="bold white", justify="center"),
        border_style="bright_blue",
        padding=(1, 4),
    ))


def render_stats(active_connections: int, total_hits: int, rotations: int) -> None:
    """Renders a quick stats bar."""
    table = Table.grid(expand=True)
    table.add_column(justify="center")
    table.add_column(justify="center")
    table.add_column(justify="center")

    table.add_row(
        Text(f"⚡ Active Connections: {active_connections}", style="bold cyan"),
        Text(f"🎯 Total Hits: {total_hits}", style="bold yellow"),
        Text(f"🔄 Rotations: {rotations}", style="bold red"),
    )

    console.print(Panel(table, border_style="dim white"))