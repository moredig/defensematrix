import sys
import json
import subprocess
from datetime import datetime


class LogPipe:
    """
    Reads real-time event logs piped from the Go engine
    and parses them into structured event dictionaries.
    """

    def __init__(self, log_file: str = None):
        self.log_file = log_file
        self.events = []

    def parse_line(self, line: str) -> dict:
        """Parses a raw [WAMAI] log line into a structured event."""
        line = line.strip()
        timestamp = datetime.now().strftime("%H:%M:%S")

        # Determine event type from log content
        if "nuked" in line.lower() or "rotation" in line.lower():
            level = "RED"
        elif "infiltrated" in line.lower() or "malicious" in line.lower():
            level = "ORANGE"
        elif "probing" in line.lower() or "suspicious" in line.lower() or "hit" in line.lower():
            level = "YELLOW"
        elif "launched" in line.lower() or "online" in line.lower() or "live" in line.lower():
            level = "BLUE"
        else:
            level = "INFO"

        event = {
            "timestamp": timestamp,
            "level": level,
            "message": line,
        }

        self.events.append(event)
        return event

    def read_from_file(self, path: str):
        """Reads and parses log events from a file."""
        try:
            with open(path, "r") as f:
                for line in f:
                    if "[WAMAI]" in line:
                        yield self.parse_line(line)
        except FileNotFoundError:
            yield {
                "timestamp": datetime.now().strftime("%H:%M:%S"),
                "level": "INFO",
                "message": "[WAMAI] Waiting for engine...",
            }

    def read_from_stdin(self):
        """Reads and parses log events from stdin pipe."""
        for line in sys.stdin:
            if "[WAMAI]" in line:
                yield self.parse_line(line)