(()=>{
const $=(s,r=document)=>r.querySelector(s);
const esc=s=>String(s).replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]));
function toast(m){let t=$('#c444-toast');if(!t){t=document.createElement('div');t.id='c444-toast';document.body.appendChild(t)}
 t.textContent=m;t.style.opacity=1;clearTimeout(t._h);t._h=setTimeout(()=>t.style.opacity=0,2500)}

/* ---- hide picker: Alt+Click hides an element, Alt+Shift+Z undoes ---- */
const KEY='444.hidden';let hidden=[];
try{hidden=JSON.parse(localStorage.getItem(KEY)||'[]')}catch(e){}
const hs=document.createElement('style');document.head.appendChild(hs);
const saveHidden=()=>{hs.textContent=hidden.map(s=>s+'{display:none!important}').join('\n');try{localStorage.setItem(KEY,JSON.stringify(hidden))}catch(e){}};
saveHidden();
function sel(el){
 const path=[];
 while(el&&el.nodeType===1&&el!==document.body){
  let s=el.tagName.toLowerCase();
  const t=el.getAttribute('data-testid'),a=el.getAttribute('aria-label');
  if(el.id){path.unshift('#'+CSS.escape(el.id));break}
  if(t){path.unshift(s+'[data-testid="'+t+'"]');break}
  if(a)s+='[aria-label="'+a.replace(/"/g,'\\"')+'"]';
  else s+=':nth-child('+([...el.parentNode.children].indexOf(el)+1)+')';
  path.unshift(s);el=el.parentElement;
 }
 return path.join('>');
}
addEventListener('click',e=>{
 if(!e.altKey||e.shiftKey)return;
 e.preventDefault();e.stopPropagation();
 const s=sel(e.target);if(!s)return;
 hidden.push(s);saveHidden();
 toast('Hidden '+document.querySelectorAll(s).length+' item(s). Alt+Shift+Z to undo');
},true);

/* ---- now playing + lyrics overlay: Alt+L, Esc or X to close (lyrics from lrclib.net) ---- */
const svg=d=>'<svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor"><path d="'+d+'"/></svg>';
const IC={sh:'M10.59 9.17 5.41 4 4 5.41l5.17 5.17 1.42-1.41zM14.5 4l2.04 2.04L4 18.59 5.41 20 17.96 7.46 20 9.5V4h-5.5zm.33 9.41-1.41 1.41 3.13 3.13L14.5 20H20v-5.5l-2.04 2.04-3.13-3.13z',pv:'M6 6h2v12H6zm3.5 6 8.5 6V6z',nx:'M6 18l8.5-6L6 6v12zM16 6v12h2V6h-2z',pl:'M8 5v14l11-7z',pa:'M6 19h4V5H6v14zm8-14v14h4V5h-4z',rp:'M7 7h10v3l4-4-4-4v3H5v6h2V7zm10 10H7v-3l-4 4 4 4v-3h12v-6h-2v4z'};
function cookie(n,a){const p=[];for(let i=0;i<120;i++){const t=i/120*Math.PI*2,r=50*(1-a*(1-Math.cos(n*t))/2);p.push((50+r*Math.cos(t)).toFixed(2)+'% '+(50+r*Math.sin(t)).toFixed(2)+'%')}return 'polygon('+p.join(',')+')'}
const secs=t=>{const p=(t||'').split(':').map(Number).reverse();return(p[0]||0)+(p[1]||0)*60+(p[2]||0)*3600};
const fmt=s=>Math.floor(s/60)+':'+String(Math.floor(s%60)).padStart(2,'0');
const txt=s=>{const e=$(s);return e?e.textContent.trim():''};
const tid=n=>$('[data-testid="'+n+'"]');
const track=()=>({title:txt('[data-testid="context-item-link"]'),artist:txt('[data-testid="context-item-info-artist"]'),
 pos:secs(txt('[data-testid="playback-position"]')),dur:secs(txt('[data-testid="playback-duration"]')),
 cover:(tid('cover-art-image')||$('[data-testid="now-playing-widget"] img')||{}).src||''});
const ov=document.createElement('div');ov.id='c444-ly';
ov.innerHTML='<div class="bg"></div><button class="x">&#10005;</button><div class="wrap"><div class="left"><div class="art"><img></div><div class="ti"></div><div class="ar"></div>'+
 '<div class="pg"><span class="cu">0:00</span><div class="bar"><i></i></div><span class="du">0:00</span></div>'+
 '<div class="ct"><button data-c="shuffle">'+svg(IC.sh)+'</button><button data-c="skip-back">'+svg(IC.pv)+'</button><button data-c="playpause" id="c444-pp">'+svg(IC.pl)+'</button><button data-c="skip-forward">'+svg(IC.nx)+'</button><button data-c="repeat">'+svg(IC.rp)+'</button></div></div><div class="ln"></div></div>';
document.body.appendChild(ov);
const ln=$('.ln',ov),ppb=$('#c444-pp',ov),art=$('.art img',ov);
art.style.clipPath=cookie(12,.07);ppb.style.clipPath=cookie(8,.12);
ov.querySelectorAll('.ct button').forEach(b=>b.onclick=()=>{const e=tid('control-button-'+b.dataset.c);if(e)e.click()});
$('.x',ov).onclick=()=>show(false);
let lines=[],key='',loaded='',cur=-1,open=false,last=-1,t0=0,lp=null;
async function load(t){
 loaded=key;lines=[];cur=-1;ln.innerHTML='<p class="m">Loading lyrics…</p>';
 try{
  const q=new URLSearchParams({track_name:t.title,artist_name:t.artist});
  const j=await(await fetch('https://lrclib.net/api/search?'+q)).json();
  const syn=j.filter(x=>x.syncedLyrics).sort((a,b)=>Math.abs(a.duration-t.dur)-Math.abs(b.duration-t.dur))[0];
  if(syn){
   lines=syn.syncedLyrics.split('\n').map(l=>{const m=l.match(/^\[(\d+):(\d+(?:\.\d+)?)\](.*)/);
    return m?{t:Number(m[1])*60+Number(m[2]),x:m[3].trim()}:null}).filter(Boolean);
   ln.innerHTML=lines.map(l=>'<p>'+(esc(l.x)||'♪')+'</p>').join('');
  }else if(j[0]&&j[0].plainLyrics){
   ln.innerHTML=j[0].plainLyrics.split('\n').map(x=>'<p class="plain">'+esc(x)+'</p>').join('');
  }else ln.innerHTML='<p class="m">No lyrics found.</p>';
 }catch(e){ln.innerHTML='<p class="m">Couldn\'t load lyrics ('+esc(e.message)+').</p>'}
}
function show(on){
 open=on;ov.classList.toggle('open',on);
 if(!on)return;
 const t=track(),big=t.cover.replace(/ab67616d0000(4851|1e02)/,'ab67616d0000b273');
 $('.bg',ov).style.backgroundImage=big?'url("'+big+'")':'none';
 art.onerror=()=>{if(art.src!==t.cover)art.src=t.cover};art.src=big;
 $('.ti',ov).textContent=t.title;$('.ar',ov).textContent=t.artist;
 if(t.title&&loaded!==t.title+'|'+t.artist)load(t);
}
addEventListener('keydown',e=>{
 if(e.altKey&&e.shiftKey&&e.code==='KeyZ'){hidden.pop();saveHidden();toast('Undone')}
 else if(e.altKey&&!e.shiftKey&&e.code==='KeyL'){e.preventDefault();show(!open)}
 else if(e.code==='Escape'&&open)show(false);
});
addEventListener('click',e=>{
 if(e.altKey||!e.target.closest)return;
 if(e.target.closest('[data-testid="lyrics-button"],button[aria-label="Lyrics"]')){e.preventDefault();e.stopPropagation();show(!open)}
},true);
setInterval(()=>{
 const t=track();if(!t.title)return;
 const k=t.title+'|'+t.artist;
 if(k!==key){key=k;if(open)show(true)}
 if(!open)return;
 if(t.pos!==last){last=t.pos;t0=performance.now()}
 const pb=tid('control-button-playpause');
 const playing=pb?/pause/i.test(pb.getAttribute('aria-label')||''):true;
 const est=t.pos+(playing?Math.min((performance.now()-t0)/1000,.99):0);
 if(playing!==lp){lp=playing;ppb.innerHTML=svg(playing?IC.pa:IC.pl)}
 $('.bar i',ov).style.width=(t.dur?Math.min(100,est/t.dur*100):0)+'%';
 $('.cu',ov).textContent=fmt(est);$('.du',ov).textContent=fmt(t.dur);
 ['shuffle','repeat'].forEach(c=>{const e=tid('control-button-'+c),b=$('[data-c="'+c+'"]',ov);
  if(e&&b)b.classList.toggle('on',['true','mixed'].includes(e.getAttribute('aria-checked')))});
 if(!lines.length)return;
 let i=-1;for(let j=0;j<lines.length;j++){if(lines[j].t<=est+.15)i=j;else break}
 if(i!==cur){cur=i;[...ln.children].forEach((p,j)=>p.className=j===i?'on':j<i?'past':'');
  if(ln.children[i])ln.children[i].scrollIntoView({block:'center',behavior:'smooth'})}
},250);
})();
