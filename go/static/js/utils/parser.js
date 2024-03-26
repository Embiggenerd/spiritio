/**
 * Parses a user message starting with '/' into a work order
 * @type {import("../../types").Parser}
 */
const parser = {
    command: '',
    i: 0,
    workOrderKey: '',
    commandConfigs: {},
    commandChar: '/',
    directMessageChar: '@',
    allLettersRegex: /[a-zA-Z]/,
    alphaNumericSpecialRegex: /[A-Za-z0-9_@./#&+!$_*+-]/,

    parseUserCommand: function (command, commandConfigs) {
        this.command = command
        this.commandConfigs = commandConfigs
        this.parse()
        // this.parseDirectMessage()

        return this.commandConfigs[this.workOrderKey]
    },

    parse: function () {
        if (this.match(this.directMessageChar)) {
            this.eat(this.directMessageChar)
            this.parseDirectMessage()
            return
        }

        if (this.match(this.commandChar)) {
            this.eat(this.commandChar)
        } else {
            return
        }

        this.parseCommand(this.readWhileMatching(this.allLettersRegex))
        if (!this.commandConfigs.hasOwnProperty(this.workOrderKey)) {
            throw new Error('no such command')
        }

        this.skipWhitespace()
        const argsCount = this.parseArguments(
            this.readWhileMatching(this.alphaNumericSpecialRegex),
            0
        )
        if (argsCount !== this.commandConfigs[this.workOrderKey].args.length) {
            throw new Error(
                `wrong argument number: have ${argsCount}, want ${
                    this.commandConfigs[this.workOrderKey].args.length
                }`
            )
        }
    },

    // Find any words that are in the list of possible commands
    parseCommand: function (word) {
        this.workOrderKey = this.workOrderKey + ' ' + word
        this.workOrderKey = this.workOrderKey.trim()

        if (
            Object.keys(this.commandConfigs).includes(
                this.workOrderKey.toLowerCase()
            ) ||
            this.i === this.command.length
        ) {
            return
        }

        this.skipWhitespace()
        this.parseCommand(this.readWhileMatching(this.allLettersRegex))
    },

    parseArguments: function (arg, count) {
        this.skipWhitespace()
        if (!arg) {
            return count
        }

        const argCfg = this.commandConfigs[this.workOrderKey].args[count]
        if (count < this.commandConfigs[this.workOrderKey].args.length) {
            argCfg.value = arg
        }

        return this.parseArguments(
            this.readWhileMatching(this.alphaNumericSpecialRegex),
            count + 1
        )
    },

    parseDirectMessage() {
        this.workOrderKey = 'direct message'
        const name = this.readWhileMatching(this.alphaNumericSpecialRegex)

        this.commandConfigs[this.workOrderKey].args[0].value = name
        this.skipWhitespace()
        const text = this.readWhileMatching(/./)
        this.commandConfigs[this.workOrderKey].args[1].value = text.trimEnd()
    },

    readWhileMatching: function (regex) {
        let startIndex = this.i
        while (
            this.i < this.command.length &&
            regex.test(this.command[this.i])
        ) {
            this.i++
            if (this.i - startIndex > 255) break
        }
        return this.command.slice(startIndex, this.i)
    },

    eat: function (str) {
        if (this.match(str)) {
            this.i += str.length
        } else {
            throw new Error(`Parse error: expecting "${str}"`)
        }
    },

    match: function (str) {
        return this.command.slice(this.i, this.i + str.length) === str
    },

    skipWhitespace: function () {
        this.readWhileMatching(/[\s\n]/)
    },
}

export default () => ({ ...parser })
