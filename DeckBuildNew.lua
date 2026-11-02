-- Era of Arcane redesigned-card deck builder for Tabletop Simulator.
-- Keep DeckBuild.lua for legacy numeric card IDs. This script accepts the
-- redesigned self-describing IDs, for example BFH001, BFCL001, BFSN002, BAIN002.

local RESOURCE_URL = "https://yifeeeeei.github.io/ArcaneImages/"
local NEW_OUTPUT_URL = RESOURCE_URL .. "new_output/"
local MAIN_DECK_BACK_URL = RESOURCE_URL .. "card_back/crad_back0.png"
local HERO_BACK_URL = RESOURCE_URL .. "card_back/crad_back4.png"

local SET_DIRECTORIES = {
    B = "%E5%9F%BA%E7%A1%80%E5%8C%85" -- 基础包
}

local ELEMENT_DIRECTORIES = {
    A = "%E6%97%A0", -- 无
    F = "%E7%81%AB", -- 火
    W = "%E6%B0%B4", -- 水
    G = "%E6%B0%94", -- 气
    E = "%E5%9C%B0", -- 地
    L = "%E5%85%89", -- 光
    D = "%E6%9A%97"  -- 暗
}

local TYPE_DIRECTORIES = {
    H = "%E4%BA%BA%E7%89%A9", -- 人物
    C = "%E4%BC%99%E4%BC%B4", -- 伙伴
    S = "%E6%8A%80%E8%83%BD", -- 技能
    I = "%E9%81%93%E5%85%B7"  -- 道具
}

local TYPE_TAGS = {
    H = "hero",
    C = "companion",
    S = "skill",
    I = "item"
}

local deckCode = ""
local builderPosition = {x = -2, y = 1, z = 3}
local buttonPosition = {x = -2, y = 1, z = -2}

function onLoad()
    setUpDeckBuilderInput()
    setUpDeckBuilderButton()
    print("新版组卡器已加载。输入新版卡号后点击按钮生成牌组。")
end

function setUpDeckBuilderInput()
    local notecard = spawnObject({
        type = "notecard",
        position = builderPosition,
        scale = {x = 0.5, y = 0.5, z = 0.5}
    })
    notecard.setLock(true)
    notecard.setDescription(" ")
    notecard.createInput({
        position = {x = 0, y = 13, z = 0},
        width = 700,
        height = 470,
        function_owner = self,
        input_function = "onDeckCodeChanged",
        interactable = true,
        readOnly = false,
        label = "输入新版牌组代码，如 BFH001 2xBFCN003 // BFSN002",
        font_size = 42
    })
end

function setUpDeckBuilderButton()
    local board = spawnObject({
        type = "Custom_Board",
        position = buttonPosition,
        scale = {x = 0.25, y = 0.5, z = 0.25}
    })
    board.setCustomObject({
        image = RESOURCE_URL .. "tts/fountain.jpg",
        thickness = 0.1,
        merge_distance = 1
    })
    board.setLock(true)
    board.createButton({
        click_function = "onBuildDeckClicked",
        function_owner = self,
        label = "生成新版牌组",
        position = {x = 0, y = 0.6, z = 0},
        rotation = {0, 0, 0},
        tooltip = "从 ArcaneImages/new_output 加载新版卡图",
        width = 8000,
        height = 8000,
        font_size = 40,
        color = {0, 0, 0, 0},
        hover_color = {1, 1, 1, 0.2}
    })
end

function onDeckCodeChanged(obj, playerColor, value, stillEditing)
    deckCode = value
end

function onBuildDeckClicked(obj, playerColor, altClick)
    buildDeckByCode(deckCode)
end

function normalizeCardId(value)
    return string.upper((value or ""):gsub("^%s+", ""):gsub("%s+$", ""))
end

function isValidCardId(cardId)
    local cardType = cardId:sub(3, 3)
    local formatIsValid = false
    if cardType == "H" then
        formatIsValid = #cardId == 6 and cardId:match("^[A-Z][A-Z]H%d%d%d$") ~= nil
    else
        formatIsValid = #cardId == 7 and cardId:match("^[A-Z][A-Z][CSI][LND]%d%d%d$") ~= nil
    end
    if not formatIsValid then
        return false
    end
    return SET_DIRECTORIES[cardId:sub(1, 1)] ~= nil
        and ELEMENT_DIRECTORIES[cardId:sub(2, 2)] ~= nil
        and TYPE_DIRECTORIES[cardId:sub(3, 3)] ~= nil
end

function parseCardToken(token)
    local count, cardId = token:match("^(%d+)[xX](.+)$")
    if cardId == nil then
        count = 1
        cardId = token
    else
        count = tonumber(count)
    end
    cardId = normalizeCardId(cardId)
    if count == nil or count < 1 or count > 99 then
        return nil, 0
    end
    return cardId, count
end

function getCardFaceUrl(cardId)
    local setDirectory = SET_DIRECTORIES[cardId:sub(1, 1)]
    local elementDirectory = ELEMENT_DIRECTORIES[cardId:sub(2, 2)]
    local typeDirectory = TYPE_DIRECTORIES[cardId:sub(3, 3)]
    return NEW_OUTPUT_URL
        .. setDirectory .. "/"
        .. typeDirectory .. "/"
        .. elementDirectory .. "/"
        .. cardId .. ".jpg"
end

function getCardBackUrl(cardId)
    if cardId:sub(3, 3) == "H" then
        return HERO_BACK_URL
    end
    return MAIN_DECK_BACK_URL
end

function getCardTag(cardId)
    return TYPE_TAGS[cardId:sub(3, 3)] or "unknown"
end

function spawnCard(cardId, position)
    local card = spawnObject({
        type = "CardCustom",
        position = position
    })
    card.setCustomObject({
        face = getCardFaceUrl(cardId),
        back = getCardBackUrl(cardId),
        width = 1,
        height = 1,
        number = 1
    })
    card.setName(cardId)
    card.setDescription("新版卡号：" .. cardId)
    card.addTag(getCardTag(cardId))
    return card
end

function spawnPile(cardIds, position)
    for index, cardId in ipairs(cardIds) do
        spawnCard(cardId, {
            x = position.x,
            y = position.y + index * 0.03,
            z = position.z
        })
    end
end

function printDeckSummary(allCardIds)
    local heroCount = 0
    local mainCount = 0
    for _, cardId in ipairs(allCardIds) do
        if cardId:sub(3, 3) == "H" then
            heroCount = heroCount + 1
        else
            mainCount = mainCount + 1
        end
    end
    print("新版牌组：人物 " .. heroCount .. " 张，主牌 " .. mainCount .. " 张。")
    if heroCount ~= 1 or mainCount ~= 40 then
        print("提示：标准新版牌组应为 1 张人物牌和 40 张主牌；当前仍会生成，便于测试。")
    end
end

function buildDeckByCode(value)
    local piles = {{}}
    local allCardIds = {}
    local invalidTokens = {}

    for token in (value or ""):gmatch("%S+") do
        if token == "//" then
            if #piles[#piles] > 0 then
                table.insert(piles, {})
            end
        else
            local cardId, count = parseCardToken(token)
            if cardId == nil or not isValidCardId(cardId) then
                table.insert(invalidTokens, token)
            else
                for copy = 1, count do
                    table.insert(piles[#piles], cardId)
                    table.insert(allCardIds, cardId)
                end
            end
        end
    end

    if #invalidTokens > 0 then
        print("无法识别的新版卡号：" .. table.concat(invalidTokens, ", "))
    end
    if #allCardIds == 0 then
        print("没有可生成的新版卡牌。")
        return
    end

    printDeckSummary(allCardIds)
    local pileIndex = 0
    for _, pile in ipairs(piles) do
        if #pile > 0 then
            pileIndex = pileIndex + 1
            spawnPile(pile, {
                x = buttonPosition.x - 6 + pileIndex * 4,
                y = buttonPosition.y + 1,
                z = buttonPosition.z + 8
            })
        end
    end
end
