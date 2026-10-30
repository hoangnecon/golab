import { useEffect, useRef } from 'react';

const RETRY_MS = 3000;

/**
 * Subscribes to a server-sent-event stream. Frames must be JSON; anything else
 * is ignored. Reconnects when the browser gives up (readyState CLOSED).
 */
export function useEventStream(
  url: string,
  onMessage: (data: unknown) => void,
  onStatus: (online: boolean) => void,
) {
  const messageRef = useRef(onMessage);
  const statusRef = useRef(onStatus);

  useEffect(() => {
    messageRef.current = onMessage;
    statusRef.current = onStatus;
  });

  useEffect(() => {
    let source: EventSource | null = null;
    let retry: ReturnType<typeof setTimeout> | undefined;
    let disposed = false;

    const connect = () => {
      source = new EventSource(url);
      source.onopen = () => statusRef.current(true);
      source.onmessage = (event: MessageEvent<string>) => {
        try {
          messageRef.current(JSON.parse(event.data));
        } catch {
          // Not a JSON frame.
        }
      };
      source.onerror = () => {
        statusRef.current(false);
        if (source?.readyState === EventSource.CLOSED && !disposed) {
          retry = setTimeout(connect, RETRY_MS);
        }
      };
    };

    connect();
    return () => {
      disposed = true;
      clearTimeout(retry);
      source?.close();
    };
  }, [url]);
}
