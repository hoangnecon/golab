export const tokens = {
  layout: {
    sidebarWidth: 'w-60',
    sidebarCollapsedWidth: 'w-16',
    headerHeight: 'h-14',
    contentMaxWidth: 'max-w-6xl',
  },
  colors: {
    canvas: 'bg-[#09090b]',
    surface: 'bg-[#121215]',
    panel: 'bg-[#18181b]',
    subtle: 'bg-zinc-900/60',
    border: 'border-zinc-800/80',
    borderHover: 'hover:border-zinc-700',
    textPrimary: 'text-zinc-100',
    textSecondary: 'text-zinc-400',
    textMuted: 'text-zinc-500',
  },
  status: {
    idle: {
      dot: 'bg-emerald-500',
      text: 'text-emerald-400',
      badge: 'bg-emerald-950/40 text-emerald-400 border-emerald-800/40',
    },
    busy: {
      dot: 'bg-amber-400 animate-pulse',
      text: 'text-amber-400',
      badge: 'bg-amber-950/40 text-amber-400 border-amber-800/40',
    },
    error: {
      dot: 'bg-red-500',
      text: 'text-red-400',
      badge: 'bg-red-950/40 text-red-400 border-red-800/40',
    },
    unknown: {
      dot: 'bg-zinc-600',
      text: 'text-zinc-500',
      badge: 'bg-zinc-900 text-zinc-400 border-zinc-800',
    },
  },
} as const;
