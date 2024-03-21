import render from './entities/renderer.js'
import media from './entities/media.js'
import message from './entities/message.js'
import parse from './entities/parse.js'
import appComponent from './component.js'

appComponent.init(render, message, media, parse)
