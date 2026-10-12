import { useState, useEffect } from 'react';

export function useSidebar() {
  const [isCollapsed, setIsCollapsed] = useState<boolean>(() => {
    const saved = localStorage.getItem('golab_sidebar_collapsed');
    if (saved !== null) {
      return saved === 'true';
    }
    // Default to collapsed if viewport is narrower than 1024px (e.g. In-App Browser)
    return typeof window !== 'undefined' ? window.innerWidth < 1024 : false;
  });

  const toggleSidebar = () => {
    setIsCollapsed((prev) => {
      const next = !prev;
      localStorage.setItem('golab_sidebar_collapsed', String(next));
      return next;
    });
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
        e.preventDefault();
        toggleSidebar();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return { isCollapsed, toggleSidebar };
}
