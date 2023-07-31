'use client'

import { CacheProvider } from '@emotion/react'
import { useEmotionCache, MantineProvider } from '@mantine/core'
import { useServerInsertedHTML } from 'next/navigation'
import React from 'react'

export default function RootStyleRegistry({
  children,
}: {
  children: React.ReactNode
}): React.JSX.Element {
  const cache = useEmotionCache()
  cache.compat = true

  useServerInsertedHTML(() => (
    <style
      dangerouslySetInnerHTML={{
        // eslint-disable-next-line @typescript-eslint/naming-convention,xss/no-mixed-html
        __html: Object.values(cache.inserted).join(' '),
      }}
      data-emotion={`${cache.key} ${Object.keys(cache.inserted).join(' ')}`}
    />
  ))

  return (
    <CacheProvider value={cache}>
      <MantineProvider withGlobalStyles withNormalizeCSS>
        {children}
      </MantineProvider>
    </CacheProvider>
  )
}
