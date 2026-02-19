import { Component, output, model } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

function describeCron(expr: string): string {
  const parts = expr.trim().split(/\s+/);
  if (parts.length !== 5) return 'Invalid expression — must have 5 parts (min hour dom month dow)';

  const [minute, hour, dom, month, dow] = parts;

  const months: Record<string, string> = {
    '1': 'January', '2': 'February', '3': 'March', '4': 'April',
    '5': 'May', '6': 'June', '7': 'July', '8': 'August',
    '9': 'September', '10': 'October', '11': 'November', '12': 'December',
  };
  const days: Record<string, string> = {
    '0': 'Sunday', '1': 'Monday', '2': 'Tuesday', '3': 'Wednesday',
    '4': 'Thursday', '5': 'Friday', '6': 'Saturday',
  };

  const timeStr =
    minute === '*' && hour === '*'
      ? 'every minute'
      : hour === '*'
      ? `at minute ${minute} of every hour`
      : minute === '*'
      ? `every minute during hour ${hour}`
      : `at ${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`;

  let dayStr = '';
  if (dom !== '*' && dow !== '*') {
    dayStr = ` on day ${dom} and every ${days[dow] ?? 'day ' + dow}`;
  } else if (dom !== '*') {
    dayStr = ` on day ${dom} of the month`;
  } else if (dow !== '*') {
    const dowLabel = dow.split(',').map((d) => days[d] ?? d).join(', ');
    dayStr = ` every ${dowLabel}`;
  }

  const monthStr =
    month !== '*'
      ? ` in ${month.split(',').map((m) => months[m] ?? m).join(', ')}`
      : '';

  return `Runs ${timeStr}${dayStr}${monthStr}.`;
}

@Component({
  selector: 'app-cron-input',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './cron-input.html',
  styleUrls: ['./cron-input.css'],
})
export class CronInputComponent {
  value = model<string>('0 * * * *');
  cronChange = output<string>();

  get humanReadable(): string {
    return describeCron(this.value());
  }

  get isValid(): boolean {
    return this.value().trim().split(/\s+/).length === 5;
  }

  onInput(val: string): void {
    this.value.set(val);
    this.cronChange.emit(val);
  }
}