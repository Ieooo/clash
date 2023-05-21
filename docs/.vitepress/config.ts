import { defineConfig } from 'vitepress'
import directoryTree from 'directory-tree'
import fs from 'fs'
import metadataParser from 'markdown-yaml-metadata-parser'

function getMetadataFromDoc(path: string): { sidebarTitle?: string, sidebarOrder?: number } {
  const fileContents = fs.readFileSync(path, 'utf8')

  return metadataParser(fileContents).metadata
}

function generateSidebarChapter(chapterDirName: string): any {
  const chapterPath = `./${chapterDirName}`
  const tree = directoryTree(chapterPath)

  if (!tree || !tree.children) {
    console.error(tree)
    throw new Error(`Could not genereate sidebar: invalid chapter at ${chapterPath}`)
  }

  let items: { sidebarOrder: number, text: string, link: string }[] = []

  // Look into files in the chapter
  for (const doc of tree.children) {
    // make sure it's a .md file
    if (doc.children || !doc.name.endsWith('.md'))
      continue

    const { sidebarOrder, sidebarTitle } = getMetadataFromDoc(doc.path)

    if (!sidebarOrder)
      throw new Error('Cannot find sidebarOrder in doc metadata: ' + doc.path)

    if (!sidebarTitle)
      throw new Error('Cannot find sidebarTitle in doc metadata: ' + doc.path)

    items.push({
      sidebarOrder,
      text: sidebarTitle,
      link: "/" + doc.path
    })
  }

  items = items.sort((a, b) => a.sidebarOrder - b.sidebarOrder)

  // remove dash and capitalize first character of each word as chapter title
  const text = chapterDirName.split('-').join(' ').replace(/\b\w/g, l => l.toUpperCase())

  return {
    text,
    collapsed: false,
    items,
  }
}

const chapters = [
  'introduction',
  'configuration',
  'premium',
  'runtime',
  'advanced-usages',
].map(generateSidebarChapter)

// Override index page link
chapters[0]['items'][0]['link'] = '/'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: 'Clash',
  description: 'Rule-based Tunnel',

  base: '/clash/',

  head: [
    [
      'link',
      { rel: 'icon', type: "image/x-icon", href: '/clash/logo.png' }
    ],
  ],

  themeConfig: {
    outline: 'deep',

    search: {
      provider: 'local'
    },

    editLink: {
      pattern: 'https://github.com/Dreamacro/clash/edit/master/docs/:path',
      text: 'Edit this page on GitHub'
    },

    logo: '/logo.png',

    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Configuration', link: '/configuration/configuration-reference' },
      {
        text: 'Download',
        items: [
          { text: 'Open-source Edition', link: 'https://github.com/Dreamacro/clash/releases/' },
          { text: 'Premium Edition', link: 'https://github.com/Dreamacro/clash/releases/tag/premium' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/Dreamacro/clash' },
    ],

    sidebar: chapters
  }
})
