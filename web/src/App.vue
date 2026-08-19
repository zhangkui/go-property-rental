<script setup lang="ts">
import {computed,onMounted} from 'vue'
import {useRoute,useRouter} from 'vue-router'
import {useAuth} from './stores/auth'
const auth=useAuth(),route=useRoute(),router=useRouter()
const loginPage=computed(()=>['/login','/register'].includes(route.path))
const menus=computed(()=>[
 {title:'运营总览',items:[{to:'/',icon:'▦',label:'仪表盘',permission:'dashboard:read'},{to:'/reports',icon:'▥',label:'经营报表',permission:'report:read'}]},
 {title:'租赁管理',items:[{to:'/properties',icon:'⌂',label:'房源单元',permission:'property:read'},{to:'/facilities',icon:'◇',label:'房源设施',permission:'facility:read'},{to:'/tenants',icon:'♙',label:'租客档案',permission:'tenant:read'},{to:'/leases',icon:'▤',label:'租约合同',permission:'lease:read'}]},
 {title:'财务管理',items:[{to:'/bills',icon:'￥',label:'账单收款',permission:'billing:read'},{to:'/billing-plans',icon:'计',label:'账单计划',permission:'billing:read'},{to:'/deposits',icon:'◉',label:'押金流水',permission:'deposit:read'},{to:'/settlements',icon:'✓',label:'退租结算',permission:'settlement:read'}]},
 {title:'运营支持',items:[{to:'/work-orders',icon:'⚒',label:'维修工单',permission:'maintenance:read'},{to:'/approvals',icon:'审',label:'审批中心',permission:'approval:read'},{to:'/notifications',icon:'铃',label:'通知提醒',permission:'notification:read'},{to:'/audits',icon:'◎',label:'审计日志',permission:'audit:read'}]},
 {title:'系统管理',items:[{to:'/users',icon:'人',label:'用户管理',permission:'user:read'},{to:'/roles',icon:'角',label:'角色管理',permission:'role:read'},{to:'/permissions',icon:'权',label:'权限管理',permission:'permission:read'},{to:'/profile',icon:'我',label:'个人资料',permission:''}]}
].map(group=>({...group,items:group.items.filter(item=>!item.permission||!auth.user||auth.hasPermission(item.permission))})).filter(group=>group.items.length))
async function logout(){await auth.logout();router.push('/login')}
onMounted(()=>auth.load().catch(()=>auth.logout()))
</script>
<template>
 <RouterView v-if="loginPage"/>
 <div v-else class="admin-shell">
  <aside class="sidebar">
   <div class="brand"><div class="brand-mark">寓</div><div><strong>长租云管家</strong><span>PROPERTY RENTAL</span></div></div>
   <div class="nav-scroll"><section v-for="group in menus" :key="group.title" class="nav-group"><p>{{group.title}}</p><RouterLink v-for="item in group.items" :key="item.to" :to="item.to" :class="{active:route.path===item.to}"><i>{{item.icon}}</i><span>{{item.label}}</span></RouterLink></section></div>
   <div class="sidebar-footer"><span class="status-dot"></span>系统服务正常</div>
  </aside>
  <section class="workspace">
   <header class="topbar"><div><h2>{{route.meta.title||'长租运营管理'}}</h2><p>集中管理房源、租约与财务运营</p></div><div class="user-area"><div class="avatar">管</div><div><strong>{{auth.user?.display_name||auth.user?.username||'管理员'}}</strong><span>{{(auth.user?.roles||[]).map((x:any)=>x.Name).join('、')||'已登录用户'}}</span></div><button class="ghost-button" @click="logout">退出登录</button></div></header>
   <div class="page-container"><RouterView/></div>
  </section>
 </div>
</template>
