// module my_address::token_transfer {
//     use aptos_framework::coin;
//     use aptos_framework::aptos_coin::AptosCoin;
//     use std::signer;
//     use aptos_std::event::{Self, EventHandle};
//     use aptos_framework::account;
    
//     // Error codes
//     const E_INSUFFICIENT_BALANCE: u64 = 1;
//     const E_CANNOT_BE_ZERO: u64 = 2;
    
//     // Transfer event structure
//     struct TransferEvent has drop, store {
//         from: address,
//         to: address,
//         amount: u64,
//     }
    
//     // Module data for storing event handlers
//     struct ModuleData has key {
//         transfer_events: EventHandle<TransferEvent>,
//     }
    
//     // Initialize module - must be private for init_module
//     fun init_module(account: &signer) {
//         if (!exists<ModuleData>(signer::address_of(account))) {
//             move_to(account, ModuleData {
//                 transfer_events: account::new_event_handle<TransferEvent>(account),
//             });
//         }
//     }
    
//     // Transfer function - Transfer a specific amount of APT to another account
//     public entry fun transfer(from: &signer, to: address, amount: u64) acquires ModuleData {
//         let from_addr = signer::address_of(from);
        
//         // Check amount is not zero
//         assert!(amount > 0, E_CANNOT_BE_ZERO);
        
//         // Check sufficient balance
//         assert!(
//             coin::balance<AptosCoin>(from_addr) >= amount,
//             E_INSUFFICIENT_BALANCE
//         );
        
//         // Execute transfer
//         coin::transfer<AptosCoin>(from, to, amount);
        
//         // Record event
//         if (exists<ModuleData>(from_addr)) {
//             let module_data = borrow_global_mut<ModuleData>(from_addr);
//             event::emit_event(&mut module_data.transfer_events, TransferEvent {
//                 from: from_addr,
//                 to,
//                 amount,
//             });
//         };
//     }
    
//     // Get account balance
//     #[view]
//     public fun get_balance(addr: address): u64 {
//         coin::balance<AptosCoin>(addr)
//     }
// } 