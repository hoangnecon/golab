import { useState, useEffect } from 'react';
import { Session, ActivityLog } from '../types';

interface SSEPayload {
  type?: string;
  sessions?: Session[];
  activity?: ActivityLog[];
}

export function useSSE(onTick?: (payload: SSEPayload) => void) {
  const [isConnected, setIsConnected] = useState<boolean>(false);

  useEffect(() => {
    let es: EventSource | null = null;
    let reconnectTimeout: NodeJS.Timeout;

    const connect = () => {
      es = new EventSource('/api/events');

      es.onopen = () => {
        setIsConnected(true);
      };

      es.onerror = () => {
        setIsConnected(false);
        es?.close();
        reconnectTimeout = setTimeout(connect, 3000);
      };

      es.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (onTick) {
            onTick(data);
          }
        } catch {
          // ignore ping messages
        }
      };
    };

    connect();

    return () => {
      clearTimeout(reconnectTimeout);
      es?.close();
    };
  }, [onTick]);

  return { isConnected };
}
