import render from './renderer.js'
import media from './entities/media.js'
import message from './entities/message.js'
import appComponent from './component.js'

let renderer = null

try {
    renderer = render()
    await appComponent.init(renderer, message, media)
} catch (e) {
    console.log(e)
    renderer.chatLog.addMessage('ADMIN (to you): ' + e)
}
