-- global variables
-- cards
local DEBUGGING = false

local card_table = {}
local resource_url = "https://yifeeeeei.github.io/ArcaneImages/"

local one_deck_capacity = 1000

-- local hero_deck_position = {
--     x = 35,
--     y = 2,
--     z = -20
-- }

-- local unit_deck_position = {
--     x = 35,
--     y = 2,
--     z = -15
-- }
-- local item_deck_position = {
--     x = 35,
--     y = 2,
--     z = 0
-- }
-- local ability_deck_position = {
--     x = 35,
--     y = 2,
--     z = 10
-- }
-- local spawn_deck_position = {
--     x = 35,
--     y = 2,
--     z = 20
-- }

-- deck builder
local deck_builder_input
local deck_builder_button
local input_str = ""

local deck_builder_input_position = {
    x = -2,
    y = 1,
    z = 3
}

local deck_builder_button_position = {
    x = -2,
    y = 1,
    z = -2
}

-- game rule
local game_rule_position = {
    x = -35, -- -35
    y = 1,
    z = 18
}

-- functions

function onLoad()
    print("onLoad")
    local jsonUrl = "https://yifeeeeei.github.io/ArcaneImages/output/simplified_card_infos.json" -- Replace this with your actual JSON URL
    -- here, I get the all card infos and then build the scene from scrach after retreiving the data
    print("setUpDeckBuilderInput")

    fetchJsonData(jsonUrl)
    -- setUpCollectionTiles()
    setUpDeckBuilder()
    -- setUpGameRule()
    -- setUpTexts()
end

-- function createTileButtonAtPostion(position, tileFrontImageUrl, tileBackImageUrl, buttonHint, onClickFunctionName)
--     local tile = spawnObject({
--         type = "Custom_Tile",
--         position = position,
--         scale = {2.4, 2.4, 2.4},
--         rotation = {0, 0, 0},
--         callback_function = function(obj)
--             obj.setLock(true)
--         end
--     })

--     tile.setCustomObject({
--         image = tileFrontImageUrl,
--         image_bottom = tileBackImageUrl,
--         thickness = 0.1, -- You can adjust the thickness of the board
--         stackable = false
--     })

--     tile.createButton({
--         click_function = onClickFunctionName,
--         function_owner = self,
--         label = buttonHint,
--         position = {0, 0.1, 0},
--         rotation = {0, 0, 0},
--         tooltip = buttonHint,
--         width = 600,
--         height = 1200,
--         font_size = 40,
--         color = {0, 0, 0, 0},
--         hover_color = {1, 1, 1, 0.2}
--     })

--     return tile
-- end

-- function setUpCollectionTiles()
--     local heroTile = createTileButtonAtPostion({
--         x = hero_deck_position.x + 6,
--         y = hero_deck_position.y - 1,
--         z = hero_deck_position.z
--     }, resource_url .. "tts/hero.jpg", resource_url .. "card_back/crad_back2.png", "英雄", "onClickHeroTile")
--     local unitTile = createTileButtonAtPostion({
--         x = unit_deck_position.x + 6,
--         y = unit_deck_position.y - 1,
--         z = unit_deck_position.z
--     }, resource_url .. "tts/unit.jpg", resource_url .. "card_back/crad_back2.png", "生物", "onClickUnitTile")
--     local itemTile = createTileButtonAtPostion({
--         x = item_deck_position.x + 6,
--         y = item_deck_position.y - 1,
--         z = item_deck_position.z
--     }, resource_url .. "tts/item.jpg", resource_url .. "card_back/crad_back2.png", "道具", "onClickItemTile")
--     local abilityTile = createTileButtonAtPostion({
--         x = ability_deck_position.x + 6,
--         y = ability_deck_position.y - 1,
--         z = ability_deck_position.z
--     }, resource_url .. "tts/ability.jpg", resource_url .. "card_back/crad_back2.png", "技能", "onClickAbilityTile")
--     local spawnTile = createTileButtonAtPostion({
--         x = spawn_deck_position.x + 6,
--         y = spawn_deck_position.y - 1,
--         z = spawn_deck_position.z
--     }, resource_url .. "tts/spawn.jpg", resource_url .. "card_back/crad_back2.png", "衍生", "onClickSpawnTile")

-- end

-- function onClickHeroTile()
--     print("onClickHeroTile")
--     local hero_card_ids = {}
--     for k, v in pairs(card_table) do
--         if isHero(k) then
--             table.insert(hero_card_ids, k)
--         end
--     end
--     if DEBUGGING then
--         hero_card_ids = {hero_card_ids[1]}
--     end
--     setUpDeck(hero_card_ids, hero_deck_position)
-- end

-- function onClickUnitTile()
--     print("onClickUnitTile")
--     local unit_card_ids = {}
--     for k, v in pairs(card_table) do
--         if isUnit(k) then
--             table.insert(unit_card_ids, k)
--         end
--     end
--     if DEBUGGING then
--         unit_card_ids = {unit_card_ids[1]}
--     end
--     setUpDeck(unit_card_ids, unit_deck_position)
-- end

-- function onClickItemTile()
--     print("onClickItemTile")
--     local item_card_ids = {}
--     for k, v in pairs(card_table) do
--         if isItem(k) then
--             table.insert(item_card_ids, k)
--         end
--     end
--     if DEBUGGING then
--         item_card_ids = {item_card_ids[1]}
--     end
--     setUpDeck(item_card_ids, item_deck_position)
-- end

-- function onClickAbilityTile()
--     print("onClickAbilityTile")
--     local ability_card_ids = {}
--     for k, v in pairs(card_table) do
--         if isAbility(k) then
--             table.insert(ability_card_ids, k)
--         end
--     end
--     if DEBUGGING then
--         ability_card_ids = {ability_card_ids[1]}
--     end
--     setUpDeck(ability_card_ids, ability_deck_position)
-- end

-- function onClickSpawnTile()
--     print("onClickSpawnTile")
--     local spawn_card_ids = {}
--     for k, v in pairs(card_table) do
--         if isSpawned(k) then
--             table.insert(spawn_card_ids, k)
--         end
--     end
--     if DEBUGGING then
--         spawn_card_ids = {spawn_card_ids[1]}
--     end
--     setUpDeck(spawn_card_ids, spawn_deck_position)
-- end

-- function setUpTexts()
--     -- deck builder
--     spawnText("组卡器", {
--         x = deck_builder_input_position.x,
--         y = deck_builder_input_position.y + 0.5,
--         z = deck_builder_input_position.z - 11
--     }, 4)
--     -- hero
--     spawnText("英雄", {
--         x = hero_deck_position.x + 3,
--         y = hero_deck_position.y - 1,
--         z = hero_deck_position.z
--     }, 2)
--     -- unit
--     spawnText("生物", {
--         x = unit_deck_position.x + 3,
--         y = unit_deck_position.y - 1,
--         z = unit_deck_position.z
--     }, 2)
--     -- item
--     spawnText("道具", {
--         x = item_deck_position.x + 3,
--         y = item_deck_position.y - 1,
--         z = item_deck_position.z
--     }, 2)
--     -- ability
--     spawnText("技能", {
--         x = ability_deck_position.x + 3,
--         y = ability_deck_position.y - 1,
--         z = ability_deck_position.z
--     }, 2)
--     -- spawn
--     spawnText("衍生", {
--         x = spawn_deck_position.x + 3,
--         y = spawn_deck_position.y - 1,
--         z = spawn_deck_position.z
--     }, 2)

-- end

-- function spawnText(text, position, scale)
--     local parameters = {
--         type = "3DText",
--         position = {
--             x = position.x,
--             y = position.y,
--             z = position.z
--         },
--         rotation = {
--             x = 90,
--             y = 180,
--             z = 0
--         },
--         scale = {
--             x = scale,
--             y = scale,
--             z = scale
--         },
--         callback_function = function(obj)
--             obj.setLock(true)
--         end

--     }
--     local text_obj = spawnObject(parameters)
--     text_obj.setValue(text)
-- end

-- function setUpGameRule()
--     local pdf_url = resource_url .. "GameRule.pdf"
--     local obj = spawnAPDF(pdf_url, "Game Rule", "Game Rule", {
--         x = 1000,
--         y = 10,
--         z = 0
--     }, 2.0)
--     obj.setPosition(game_rule_position)
-- end

-- function spawnAPDF(url, name, description, position, scale)
--     local myjson = [[
--         {
--           "Name": "Custom_PDF",
--           "Transform": {
--             "posX": ]] .. position.x .. [[,
--             "posY": ]] .. position.y .. [[,
--             "posZ": ]] .. position.z .. [[,
--             "rotX": 0.0,
--             "rotY": 0.0,
--             "rotZ": 0.0,
--             "scaleX": ]] .. scale .. [[,
--             "scaleY": ]] .. scale .. [[,
--             "scaleZ": ]] .. scale .. [[
--           },
--           "Nickname": "]] .. name .. [[",
--           "Description": "]] .. description .. [[",
--           "GMNotes": "",
--           "ColorDiffuse": {
--             "r": 1.0,
--             "g": 1.0,
--             "b": 1.0
--           },
--           "Locked": false,
--           "Grid": false,
--           "Snap": true,
--           "IgnoreFoW": false,
--           "Autoraise": true,
--           "Sticky": true,
--           "Tooltip": true,
--           "GridProjection": false,
--           "HideWhenFaceDown": false,
--           "Hands": false,
--           "CustomPDF": {
--             "PDFUrl": "]] .. url .. [[",
--             "PDFPassword": "",
--             "PDFPage": 0,
--             "PDFPageOffset": 0
--           },
--           "XmlUI": "<!-- -->",
--           "LuaScript": "--foo",
--           "LuaScriptState": "",
--           "GUID": "pdf001"
--         }]]
--     local obj = spawnObjectJSON({
--         json = myjson,
--         position = {0, 5, 0}
--     })
--     return obj
-- end

-- this function is called after fetching the json data

function fetchJsonDataCallback()
    -- print("entering fetchJsonDataCallback")
    -- setUpAllDecks()
    print("finish fetchJsonDataCallback")

end

function setUpDeckBuilder()
    print("setUpDeckBuilder")
    setUpDeckBuilderInput()
    setUpDeckBuilderButton()
end

function setUpDeckBuilderInput()
    local parameters = {
        type = "notecard",
        position = {
            x = deck_builder_input_position.x,
            y = deck_builder_input_position.y,
            z = deck_builder_input_position.z
        },
        scale = {
            x = 0.5,
            y = 0.5,
            z = 0.5
        },
        callback_function = function(obj)
            obj.setLock(true)
        end

    }
    local notecard = spawnObject(parameters)
    notecard.setDescription(" ")
    notecard.createInput({
        position = {
            x = 0,
            y = 3 + 10,
            z = 0
        },
        width = 500,
        height = 470,
        function_owner = self,
        input_function = "input_func",
        interactable = true,
        readOnly = false,
        label = "Input deck code here. Click the magic fountain to build the deck.",
        font_size = 50
    })

end

function input_func(obj, color, input, stillEditing)
    input_str = input
end

function setUpDeckBuilderButton()
    local parameters = {
        type = "Custom_Board",
        position = {
            x = deck_builder_button_position.x,
            y = deck_builder_button_position.y,
            z = deck_builder_button_position.z
        },
        scale = {
            x = 0.25,
            y = 0.5,
            z = 0.25
        },
        callback_function = function(obj)
            obj.setLock(true)
        end

    }
    local custom_board = spawnObject(parameters)
    -- custom_board.setLock(true)

    custom_board.setCustomObject({
        image = resource_url .. "tts/fountain.jpg",
        thickness = 0.1, -- You can adjust the thickness of the board
        merge_distance = 1 -- The distance at which the board will merge with other objects
    })

    custom_board.createButton({
        click_function = 'onClickDeckBuilderButton',
        function_owner = self,
        label = 'click me to set deck',
        position = {
            x = 0,
            y = 0.6,
            z = 0
        },
        rotation = {0, 0, 0},
        tooltip = "click me to set deck",
        width = 8000,
        height = 8000,
        font_size = 40,
        color = {0, 0, 0, 0},
        hover_color = {1, 1, 1, 0.2}
    })

end

function onClickDeckBuilderButton()
    print("onClickDeckBuilderButton")
    buildDeckByDeckCode(input_str)
end

function splitByWhitespace(str)
    local words = {}
    for word in str:gmatch("%S+") do
        table.insert(words, word)
    end
    return words
end

function buildDeckByDeckCode(deck_code)
    local words = splitByWhitespace(deck_code)
    local card_codes = {}
    local pile_created = 1
    print("number of words: ", #words)

    for i = 1, #words do
        local card_code = words[i]
        if card_code == "//" then
            print("comment: ", card_code)
            if #card_codes > 0 then
                setUpDeck(card_codes, {
                    x = deck_builder_button_position.x - 10 + pile_created * 4,
                    y = deck_builder_button_position.y,
                    z = deck_builder_button_position.z + 8
                })
                pile_created = pile_created + 1
                card_codes = {}
            end
        elseif card_table[card_code] == nil then
            print("card code not found: ", card_code)
        else
            table.insert(card_codes, card_code)
        end
    end
    if #card_codes > 0 then
        setUpDeck(card_codes, {
            x = deck_builder_button_position.x - 10 + pile_created * 4,
            y = deck_builder_button_position.y,
            z = deck_builder_button_position.z + 8
        })
        pile_created = pile_created + 1
        card_codes = {}
    end
end

function setUpDeck(card_ids, position)
    local one_deck = {}
    local z_pos = position.z
    for i = 1, #card_ids do
        table.insert(one_deck, card_ids[i])
        if #one_deck == one_deck_capacity then
            spawnDeck(one_deck, {
                x = position.x,
                y = position.y,
                z = z_pos
            }, true)
            z_pos = z_pos + 3.5
            one_deck = {}
        end
    end
    if #one_deck > 0 then
        spawnDeck(one_deck, {
            x = position.x,
            y = position.y,
            z = z_pos
        }, true)
    end

end

-- this function has been deprecated
-- function setUpAllDecks()
--     print("setUpAllDecks")
--     local hero_card_ids = {}
--     local unit_card_ids = {}
--     local item_card_ids = {}
--     local ability_card_ids = {}
--     local spawn_card_ids = {}
--     local count = 0
--     for _ in pairs(card_table) do
--         count = count + 1
--     end
--     print("all cards: ", count)

--     for k, v in pairs(card_table) do
--         if isHero(k) then
--             table.insert(hero_card_ids, k)
--         elseif isSpawned(k) then
--             table.insert(spawn_card_ids, k)
--         elseif isUnit(k) then
--             table.insert(unit_card_ids, k)
--         elseif isItem(k) then
--             table.insert(item_card_ids, k)
--         elseif isAbility(k) then
--             table.insert(ability_card_ids, k)
--         end
--     end

--     table.sort(hero_card_ids)
--     table.sort(unit_card_ids)
--     table.sort(item_card_ids)
--     table.sort(ability_card_ids)
--     table.sort(spawn_card_ids)

--     -- when testing, cut each deck to only 1 card
--     if DEBUGGING then
--         hero_card_ids = {hero_card_ids[1]}
--         unit_card_ids = {unit_card_ids[1]}
--         item_card_ids = {item_card_ids[1]}
--         ability_card_ids = {ability_card_ids[1]}
--         spawn_card_ids = {spawn_card_ids[1]}
--     end

--     print("hero cards: ", #hero_card_ids)
--     print("unit cards: ", #unit_card_ids)
--     print("item cards: ", #item_card_ids)
--     print("ability cards: ", #ability_card_ids)
--     print("spawn cards: ", #spawn_card_ids)

--     print("Setting up hero deck")
--     setUpDeck(hero_card_ids, hero_deck_position)
--     print("Setting up unit deck")
--     setUpDeck(unit_card_ids, unit_deck_position)
--     print("Setting up item deck")
--     setUpDeck(item_card_ids, item_deck_position)
--     print("Setting up ability deck")
--     setUpDeck(ability_card_ids, ability_deck_position)
--     print("Setting up spawn deck")
--     setUpDeck(spawn_card_ids, spawn_deck_position)
-- end

function spawnDeck(card_ids, position, need_tag)
    local card_face_urls = {}
    local card_back_urls = {}
    for k, v in pairs(card_ids) do
        local card_id = card_ids[k]
        local faceUrl = getCardFaceUrl(card_id)
        local backUrl = getCardBackUrl(card_id)
        table.insert(card_face_urls, faceUrl)
        table.insert(card_back_urls, backUrl)
    end
    for index = 1, #card_ids do
        local faceUrl = card_face_urls[index]
        local cardBack = card_back_urls[index]
        local card = spawnCard(faceUrl, cardBack, {
            x = position.x,
            y = position.y + index * 0.01,
            z = position.z
        })
        if need_tag then
            local tag = getTag(card_ids[index])
            card.addTag(tag)
        end
    end
end

function getTag(card_id)
    if isHero(card_id) then
        return "hero"
    elseif isUnit(card_id) then
        return "creature"
    elseif isItem(card_id) then
        return "item"
    elseif isAbility(card_id) then
        return "ability"
    else
        return "unkown"
    end
end

function isSpawned(card_id)
    -- return a bool
    return card_id:sub(3, 3) == '0'
end

function isItem(card_id)
    -- return a bool
    return card_id:sub(1, 1) == '2'
end

function isUnit(card_id)
    -- return a bool
    return card_id:sub(1, 1) == '1'
end

function isAbility(card_id)
    -- return a bool
    return card_id:sub(1, 1) == '3'
end

function isHero(card_id)
    -- return a bool
    return card_id:sub(1, 1) == '4'
end

function dump(o)
    if type(o) == 'table' then
        local s = '{ '
        for k, v in pairs(o) do
            if type(k) ~= 'number' then
                k = '"' .. k .. '"'
            end
            s = s .. '[' .. k .. '] = ' .. dump(v) .. ','
        end
        return s .. '} '
    else
        return tostring(o)
    end
end

function fetchJsonData(url)
    WebRequest.get(url, function(webRequest)
        print("Request complete")
        if webRequest.is_done then
            local data = JSON.decode(webRequest.text)
            card_table = data
            fetchJsonDataCallback()
        end
    end)
end

function getCardElement(card_id)
    local elem_number = card_id:sub(2, 2)
    if elem_number == '0' then
        return "无"
    elseif elem_number == '1' then
        return "火"
    elseif elem_number == '2' then
        return "水"
    elseif elem_number == '3' then
        return "气"
    elseif elem_number == '4' then
        return "地"
    elseif elem_number == '5' then
        return "光"
    elseif elem_number == '6' then
        return "暗"
    else
        print("error when getCardElement " .. card_id)
        return "无"
    end
end

function getCardType(card_id)
    local type_number = card_id:sub(1, 1)

    if type_number == '1' then
        return "生物"
    elseif type_number == '2' then
        return "道具"
    elseif type_number == '3' then
        return "技能"
    elseif type_number == '4' then
        return "英雄"
    else
        print("error when getCardType " .. card_id)
        return "无"
    end
end

function getCardFaceUrl(card_id)
    local url = resource_url .. card_table[card_id]
    return url
end

function getCardBackUrl(card_id)
    if isAbility(card_id) then
        return resource_url .. "card_back/" .. "crad_back3.png"
    elseif isHero(card_id) then
        return resource_url .. "card_back/" .. "crad_back4.png"
    else
        return resource_url .. "card_back/" .. "crad_back0.png"
    end
end

function spawnCard(faceUrl, backUrl, position)
    local card = spawnObject({
        type = "CardCustom",
        position = position
    })

    card.setCustomObject({
        face = faceUrl,
        back = backUrl,
        width = 1,
        height = 1,
        number = 1
    })

    return card
end

