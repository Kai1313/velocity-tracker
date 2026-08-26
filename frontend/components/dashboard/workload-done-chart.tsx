'use client';

import { Bar, BarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

export type WorkloadDoneDatum = {
  category: string;
  workload: number;
  done: number;
};

type TooltipPayloadEntry = {
  dataKey: string;
  name: string;
  value: number;
  color: string;
};

function ChartTooltip({
  active,
  payload,
  label,
}: {
  active?: boolean;
  payload?: TooltipPayloadEntry[];
  label?: string;
}) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-md border border-border bg-card px-3 py-2 text-sm shadow-md">
      <p className="mb-1 font-medium text-foreground">{label}</p>
      {payload.map((p) => (
        <div key={p.dataKey} className="flex items-center gap-2">
          <span className="h-0.5 w-3 shrink-0 rounded-full" style={{ backgroundColor: p.color }} />
          <span className="font-medium tabular-nums text-foreground">{p.value}</span>
          <span className="text-muted-foreground">{p.name}</span>
        </div>
      ))}
    </div>
  );
}

// Workload and Done are always plotted in this fixed order/color pair —
// validated colorblind-safe (CVD ΔE, normal-vision ΔE, contrast) against this
// app's actual light/dark card surfaces via the dataviz skill's palette
// validator, not eyeballed.
export function WorkloadDoneChart({ data }: { data: WorkloadDoneDatum[] }) {
  if (data.length === 0) return null;

  return (
    <div className="h-72 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }} barGap={2} barCategoryGap="20%">
          <CartesianGrid vertical={false} stroke="hsl(var(--border))" />
          <XAxis
            dataKey="category"
            stroke="hsl(var(--muted-foreground))"
            fontSize={12}
            tickLine={false}
            axisLine={{ stroke: 'hsl(var(--border))' }}
          />
          <YAxis stroke="hsl(var(--muted-foreground))" fontSize={12} tickLine={false} axisLine={false} allowDecimals={false} />
          <Tooltip content={<ChartTooltip />} cursor={{ fill: 'hsl(var(--muted))' }} />
          <Legend wrapperStyle={{ fontSize: 12, color: 'hsl(var(--muted-foreground))' }} />
          <Bar dataKey="workload" name="Workload" fill="var(--chart-workload)" radius={[4, 4, 0, 0]} barSize={24} />
          <Bar dataKey="done" name="Done" fill="var(--chart-done)" radius={[4, 4, 0, 0]} barSize={24} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
