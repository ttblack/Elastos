--- This is a test about on duty arbitrator successfully
---
local api = require("api")
local colors = require 'test/common/ansicolors'
local dpos_msg = require("test/white_box/dpos_msg")
local log = require("test/white_box/log_config")
local block_utils = require("test/white_box/block_utils")

print(colors('%{blue}-----------------Begin-----------------'))
local dpos = dofile("test/white_box/dpos_manager.lua")

api.clear_store()
api.init_ledger(log.level, dpos.A.arbitrators)

dpos.set_on_duty(1) -- set A on duty

--- initial status check
assert(dpos.A.manager:is_on_duty())
assert(dpos.A.manager:is_status_ready())
assert(not dpos.A.manager:is_status_running())

--- generate two blocks within same height
local b1 = block_utils.height_one()
local b2 = block_utils.height_one()
assert(b1:hash() ~= b2:hash())

--- simulate block arrive event
local prop = proposal.new(dpos.A.manager:public_key(), b1:hash(), 0)
dpos.A.manager:sign_proposal(prop)

local va = vote.new(prop:hash(), dpos.A.manager:public_key(), true)
dpos.A.network:push_block(b1, false)
assert(dpos.A.network:check_last_msg(dpos_msg.accept_vote, va))
dpos.A.network:push_block(b2, false)

--- simulate other arbitrators' approve votes
local vb = vote.new(prop:hash(), dpos.B.manager:public_key(), true)
dpos.B.manager:sign_vote(vb)
dpos.A.network:push_vote(dpos.B.manager:public_key(), vb)

local vc = vote.new(prop:hash(), dpos.C.manager:public_key(), true)
dpos.C.manager:sign_vote(vc)
dpos.A.network:push_vote(dpos.C.manager:public_key(), vc)

local confirm = confirm.new(b1:hash())
confirm:set_proposal(prop)
confirm:append_vote(va)
confirm:append_vote(vb)
confirm:append_vote(vc)
assert(dpos.A.manager:check_last_relay(1, confirm))

print(dpos.A.manager:dump_node_relays())
print(dpos.A.network:dump_msg())

--- clean up
api.close_store()
print(colors('%{green}-----------------Test success!-----------------'))
